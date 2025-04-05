// Copyright (C) 2025 T-Force I/O
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
	"errors"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx/slicext"
	"github.com/tforce-io/tf-golib/strfmt"
	"github.com/tforceaio/tf-unifiler-go/core"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/db"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

// MetadataModule handles user requests related file hashes.
type MetadataModule struct {
	logger zerolog.Logger
}

// Return new MetadataModule.
func NewMetadataModule(c *Controller, cmdName string) *MetadataModule {
	return &MetadataModule{
		logger: c.CommandLogger("metadata", cmdName),
	}
}

// Scan and compute hashes using common algorithms (MD5, SHA-1, SHA-256, SHA-512) for inputs (files/folders)
// and add them to collection.
// Mark them as obseleted if delete is true.
func (m *MetadataModule) Scan(workspaceDir string, inputs, collections []string, delete bool) error {
	if workspaceDir == "" {
		return errors.New("workspace is not set")
	} else if !filesystem.IsDirectoryExist(workspaceDir) {
		return errors.New("workspace is not found")
	}
	if len(inputs) == 0 {
		return errors.New("inputs is empty")
	}
	if len(collections) == 0 {
		return errors.New("collections is empty")
	}
	m.logger.Info().
		Strs("collections", collections).
		Bool("delete", delete).
		Strs("files", inputs).
		Str("workspace", workspaceDir).
		Msg("Start scanning files metadata.")

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

	dbFile := MetadataWorkspaceDatabase(workspaceDir)
	ctx, err := db.Connect(dbFile)
	if err != nil {
		return err
	}
	err = m.saveHResults(ctx, hResults, delete, collections)
	if err != nil {
		return err
	}

	return nil
}

// Decorator to log error occurred when calling handlers.
func (m *MetadataModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
}

// Save hashing results to metadata database along with their respective collections.
func (m *MetadataModule) saveHResults(ctx *db.DbContext, hResults []*core.FileMultiHash, ignore bool, collections []string) (err error) {
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

// Return database path to store Metadata module's ouputs inside Unifiler workspace.
func MetadataWorkspaceDatabase(workspaceDir string) string {
	return filepath.Join(workspaceDir, ".unifiler", "metadata.db")
}

// Define Cobra Command for Metadata module.
func MetadataCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "metadata",
		Short: "File metadata management and mass file update using metadata.",
	}

	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Compute hashes for files using common algorithms (MD5, SHA-1, SHA-256, SHA-512) and persist them.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseMetadataFlags(cmd)
			m := NewMetadataModule(c, "scan")
			m.logError(m.Scan(flags.WorkspaceDir, flags.Inputs, flags.Collections, flags.Deleted))
		},
	}
	scanCmd.Flags().StringSliceP("collections", "c", []string{}, "Names of collections of known files. If a collection existed, files will be appended to that collection.")
	scanCmd.Flags().Bool("delete", false, "Mark the inputs as obsoleted.")
	scanCmd.Flags().StringSliceP("inputs", "i", []string{}, "Files/Directories to hash.")
	scanCmd.Flags().StringP("workspace", "w", "", "Directory contains Unifiler workspace.")
	rootCmd.AddCommand(scanCmd)

	return rootCmd
}

// Struct FileFlags contains all flags used by Metadata module.
type MetadataFlags struct {
	Collections  []string
	Deleted      bool
	Inputs       []string
	WorkspaceDir string
}

// Extract all flags from a Cobra Command.
func ParseMetadataFlags(cmd *cobra.Command) *MetadataFlags {
	collections, _ := cmd.Flags().GetStringSlice("collections")
	deleted, _ := cmd.Flags().GetBool("deleted")
	inputs, _ := cmd.Flags().GetStringSlice("inputs")
	workspaceDir, _ := cmd.Flags().GetString("workspace")

	return &MetadataFlags{
		Collections:  collections,
		Deleted:      deleted,
		Inputs:       inputs,
		WorkspaceDir: workspaceDir,
	}
}
