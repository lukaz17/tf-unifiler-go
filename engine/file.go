// Copyright (C) 2024 T-Force I/O
// This file is part of TF Unifiler
//
// TF Unifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TF Unifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TF Unifiler. If not, see <https://www.gnu.org/licenses/>.

package engine

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforce-io/tf-golib/opx/slicext"
	"github.com/tforce-io/tf-golib/strfmt"
	"github.com/tforceaio/tf-unifiler-go/core"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/db"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

// Struct FileRenameMapping stores old and new filename after renaming for rollback.
type FileRenameMapping struct {
	Source string `json:"s,omitempty"`
	Target string `json:"t,omitempty"`
}

// FileModule handles user requests related to batch processing of files in general.
type FileModule struct {
	logger zerolog.Logger
}

// Return new FileModule.
func NewFileModule(c *Controller, cmdName string) *FileModule {
	return &FileModule{
		logger: c.CommandLogger("file", cmdName),
	}
}

// Scan and compute hashes using common algorithms (MD5, SHA-1, SHA-256, SHA-512) for inputs (files/folders).
// Mark them as obseleted if delete is true.
func (m *FileModule) Hash(workspaceDir string, inputs, collections []string, delete bool) error {
	if workspaceDir == "" {
		return errors.New("workspace is not set")
	} else if !filesystem.IsDirectoryExist(workspaceDir) {
		return errors.New("workspace is not found")
	}
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}
	m.logger.Info().
		Strs("collections", collections).
		Bool("delete", delete).
		Strs("files", inputs).
		Msg("Start hashing files.")

	contents, err := filesystem.List(inputs, true)
	if err != nil {
		return err
	}

	hResults := []*core.FileMultiHash{}
	algos := []string{"md5", "sha1", "sha256", "sha512"}
	for _, c := range contents {
		if c.IsDir {
			continue
		}
		fhResults, err := hasher.Hash(c.RelativePath, algos)
		if err != nil {
			m.logger.Info().
				Str("path", c.RelativePath).
				Msg("Failed to compute hash.")
			return err
		}
		m.logger.Info().
			Strs("algos", algos).
			Str("path", c.RelativePath).
			Int("size", fhResults[0].Size).
			Msg("Hashed file.")
		fileMultiHash := &core.FileMultiHash{
			Md5:      fhResults[0].Hash,
			Sha1:     fhResults[1].Hash,
			Sha256:   fhResults[2].Hash,
			Sha512:   fhResults[3].Hash,
			Size:     uint32(fhResults[0].Size),
			FileName: c.Name,
		}
		hResults = append(hResults, fileMultiHash)
	}

	dbFile := FileWorkspaceDatabase(workspaceDir)
	ctx, err := db.Connect(dbFile)
	if err != nil {
		return err
	}
	err = m.saveHResults(ctx, hResults, delete, collections)
	if err != nil {
		return err
	}

	if delete {
		for _, c := range contents {
			if c.IsDir {
				continue
			}
			err = os.Remove(c.AbsolutePath)
			if err != nil {
				m.logger.Info().
					Str("path", c.RelativePath).
					Msg("Failed to delete file.")
				return err
			}
			m.logger.Info().
				Str("path", c.RelativePath).
				Msg("Deleted file.")
		}
	}

	return nil
}

// Multi-rename files. Input which is directories will be ignored.
func (m *FileModule) Rename(inputs []string, preset string) error {
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}
	m.logger.Info().
		Strs("inputs", inputs).
		Str("preset", preset).
		Msg("Start renaming file.")

	if preset == "md4" {
		return m.renameByHash(inputs, preset, "6d6434_")
	}
	if preset == "md5" {
		return m.renameByHash(inputs, preset, "6d6435_")
	}
	if preset == "sha1" {
		return m.renameByHash(inputs, preset, "73686131_")
	}
	if preset == "sha256" {
		return m.renameByHash(inputs, preset, "736861323536_")
	}
	if preset == "sha512" {
		return m.renameByHash(inputs, preset, "736861353132_")
	}

	return errors.New("preset is invalid")
}

// Rename files using hashes of their contents.
func (m *FileModule) renameByHash(inputs []string, algo string, prefix string) error {
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}

	contents, err := filesystem.List(inputs, false)
	if err != nil {
		return err
	}

	files := []*filesystem.FsEntry{}
	for _, c := range contents {
		if !c.IsDir {
			files = append(files, c)
		}
	}
	hResults := []*hasher.HashResult{}
	for _, c := range files {
		if c.IsDir {
			continue
		}
		fhResults, err := hasher.Hash(c.RelativePath, []string{algo})
		if err != nil {
			m.logger.Info().
				Str("path", c.RelativePath).
				Msg("Failed to compute hash.")
			return err
		}
		m.logger.Info().
			Str("algo", algo).
			Str("path", c.RelativePath).
			Int("size", fhResults[0].Size).
			Msg("Hashed file.")
		hResults = append(hResults, fhResults...)
	}

	mappings := []*FileRenameMapping{}
	for _, e := range hResults {
		parent := path.Dir(e.Path)
		ext := path.Ext(e.Path)
		targetName := prefix + hex.EncodeToString(e.Hash) + ext
		target := opx.Ternary(parent == ".", targetName, filesystem.Join(parent, targetName))
		mapping := &FileRenameMapping{
			Source: e.Path,
			Target: target,
		}
		mappings = append(mappings, mapping)
	}

	currentTimestamp := time.Now().UnixMilli()
	rollbackFilePath := filesystem.Join(".", "unifiler-file-rename-"+strconv.FormatInt(currentTimestamp, 10)+".json")
	fContent, _ := json.Marshal(mappings)
	fContents := []string{string(fContent)}
	err = filesystem.WriteLines(rollbackFilePath, fContents)
	if err == nil {
		m.logger.Info().
			Int("lineCount", len(fContents)).
			Str("path", rollbackFilePath).
			Msg("Written rollback file.")
	} else {
		m.logger.Info().
			Str("path", rollbackFilePath).
			Msg("Failed to write rollback file.")
		return err
	}

	for _, e := range mappings {
		if e.Source == e.Target {
			m.logger.Info().
				Str("src", e.Source).
				Str("dest", e.Target).
				Msg("Skipped. File is already renamed.")
			continue
		}
		if filesystem.IsFileExist(e.Target) {
			m.logger.Info().
				Str("src", e.Source).
				Str("dest", e.Target).
				Msg("Skipped. Target file is existed.")
			continue
		}
		err := os.Rename(e.Source, e.Target)
		if err != nil {
			m.logger.Err(err)
			m.logger.Info().
				Str("src", e.Source).
				Str("dest", e.Target).
				Msg("Failed to rename file.")
			continue
		}
		m.logger.Info().
			Str("src", e.Source).
			Str("target", e.Target).
			Msg("Renamed file.")
	}

	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *FileModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
}

// Save hashing results to metadata database along with their respective collections.
func (m *FileModule) saveHResults(ctx *db.DbContext, hResults []*core.FileMultiHash, ignore bool, collections []string) (err error) {
	// save Hash
	hashes := make([]*db.Hash, len(hResults))
	for i, res := range hResults {
		hashes[i] = db.NewHash(res, ignore)
	}
	err = ctx.SaveHashes(hashes)
	if err != nil {
		m.logger.Info().Msg("Failed to save Hashes.")
		return err
	}
	// save Mapping
	sha256s := make([]string, len(hResults))
	for i, res := range hResults {
		sha256s[i] = res.Sha256.HexStr()
	}
	hashes, err = ctx.GetHashesBySha256s(sha256s)
	if err != nil {
		m.logger.Info().Msg("Failed to reload Hashes.")
		return err
	}
	hashesMap := map[string]uuid.UUID{}
	for _, hash := range hashes {
		hashesMap[hash.Sha256] = hash.ID
	}
	mappings := make([]*db.Mapping, len(hResults))
	for i, res := range hResults {
		fileName := strfmt.NewFileNameFromStr(res.FileName)
		mappings[i] = db.NewMapping(hashesMap[res.Sha256.HexStr()], fileName.Name, fileName.Extension)
	}
	err = ctx.SaveMappings(mappings)
	if err != nil {
		m.logger.Info().Msg("Failed to save Mappings.")
		return err
	}
	if !slicext.IsEmpty(collections) {
		// save Set
		sets := make([]*db.Set, len(collections))
		for i, name := range collections {
			sets[i] = db.NewSet(name)
		}
		err = ctx.SaveSets(sets)
		if err != nil {
			m.logger.Info().Msg("Failed to save Sets.")
			return err
		}
		// save SetHash
		sets, err = ctx.GetSetsByNames(collections)
		if err != nil {
			m.logger.Info().Msg("Failed to reload Sets.")
			return err
		}
		setHashes := make([]*db.SetHash, len(sets)*len(hashes))
		for i, set := range sets {
			hashesLen := len(hashes)
			for j, hash := range hashes {
				setHashes[i*hashesLen+j] = db.NewSetHash(set.ID, hash.ID)
			}
		}
		err = ctx.SaveSetHashes(setHashes)
		if err != nil {
			m.logger.Info().Msg("Failed to save SetHashes.")
			return err
		}
	}

	m.logger.Info().Msg("Saved metadata successfully.")
	return err
}

// Return database path to store File module's ouputs inside Unifiler workspace.
func FileWorkspaceDatabase(workspaceDir string) string {
	return filepath.Join(workspaceDir, ".unifiler", "metadata.db")
}

// Define Cobra Command for File module.
func FileCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "file",
		Short: "Batch file processing in general.",
	}

	hashCmd := &cobra.Command{
		Use:   "hash",
		Short: "Compute hashes for files using common algorithms (MD5, SHA-1, SHA-256, SHA-512).",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseFileFlags(cmd)
			m := NewFileModule(c, "hash")
			m.logError(m.Hash(flags.WorkspaceDir, flags.Inputs, flags.Collections, flags.Deleted))
		},
	}
	hashCmd.Flags().StringSliceP("collections", "c", []string{}, "Names of collections of known files. If a collection existed, files will be appended to that collection.")
	hashCmd.Flags().Bool("delete", false, "Mark the inputs as obsoleted.")
	hashCmd.Flags().StringSliceP("inputs", "i", []string{}, "Files/Directories to hash.")
	hashCmd.Flags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")
	rootCmd.AddCommand(hashCmd)

	renameCmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename multiples file using pre-defined settings.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseFileFlags(cmd)
			m := NewFileModule(c, "rename")
			m.logError(m.Rename(flags.Inputs, flags.Preset))
		},
	}
	renameCmd.Flags().StringSliceP("inputs", "i", []string{}, "Files to rename. Directories will be ignored.")
	renameCmd.Flags().StringP("preset", "p", "", "Name of pre-defined settings for renaming.")
	rootCmd.AddCommand(renameCmd)

	return rootCmd
}

// Struct FileFlags contains all flags used by File module.
type FileFlags struct {
	Collections  []string
	Deleted      bool
	Inputs       []string
	Preset       string
	WorkspaceDir string
}

// Extract all flags from a Cobra Command.
func ParseFileFlags(cmd *cobra.Command) *FileFlags {
	collections, _ := cmd.Flags().GetStringSlice("collections")
	deleted, _ := cmd.Flags().GetBool("deleted")
	inputs, _ := cmd.Flags().GetStringSlice("inputs")
	preset, _ := cmd.Flags().GetString("preset")
	workspaceDir, _ := cmd.Flags().GetString("workspace")

	return &FileFlags{
		Collections:  collections,
		Deleted:      deleted,
		Inputs:       inputs,
		Preset:       preset,
		WorkspaceDir: workspaceDir,
	}
}
