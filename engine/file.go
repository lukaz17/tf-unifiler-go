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
	"os"
	"path"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/opx/slicext"
	"github.com/tforce-io/tf-golib/strfmt"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/core"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/db"
	"github.com/tforceaio/tf-unifiler-go/extension"
	"github.com/tforceaio/tf-unifiler-go/extension/generic"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

type FileModule struct {
	logger zerolog.Logger
}

func NewFileModule(c *Controller) *FileModule {
	return &FileModule{
		logger: c.ModuleLogger("File"),
	}
}

type FileRenameMapping struct {
	Source string `json:"s,omitempty"`
	Target string `json:"t,omitempty"`
}

func (m *FileModule) File(args *cmd.FileCmd) {
	if args.Delete != nil {
		m.Scan((*cmd.FileScanCmd)(args.Delete), true)
	} else if args.Rename != nil {
		m.Rename(args.Rename)
	} else if args.Scan != nil {
		m.Scan(args.Scan, false)
	} else {
		m.logger.Error().Msg("Invalid arguments")
	}
}

func (m *FileModule) Scan(args *cmd.FileScanCmd, delete bool) {
	if len(args.Files) == 0 {
		m.logger.Error().Msg("No input file")
		return
	}
	m.logger.Info().
		Array("files", extension.StringSlice(args.Files)).
		Msgf("Start deleting files")

	contents, err := filesystem.List(args.Files, true)
	if err != nil {
		m.logger.Err(err).Msg("Error listing input files")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}

	hResults := []*core.FileMultiHash{}
	algos := []string{"md5", "sha1", "sha256", "sha512"}
	for _, c := range contents {
		if c.IsDir {
			continue
		}
		fhResults, err := hasher.Hash(c.RelativePath, algos)
		if err != nil {
			m.logger.Err(err).Msgf("Error computing hash for '%s'", c.RelativePath)
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
		m.logger.Info().Array("algos", extension.StringSlice(algos)).Int("size", fhResults[0].Size).Msgf("File hashed '%s'", c.RelativePath)
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
	dbFile := filesystem.Join(args.Workspace, "metadata.db")
	ctx, err := db.Connect(dbFile)
	if err != nil {
		m.logger.Err(err).Msg("Error while opening metadata database.")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}
	err = m.saveHResults(ctx, hResults, delete, args.Collections)
	if err != nil {
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}

	if delete {
		for _, c := range contents {
			if c.IsDir {
				continue
			}
			err = os.Remove(c.AbsolutePath)
			if err != nil {
				m.logger.Err(err).Str("Path", c.RelativePath).Msg("Error while deleting file.")
				m.logger.Info().Msg("Unexpected error occurred. Exiting...")
				return
			}
			m.logger.Info().Str("Path", c.RelativePath).Msgf("File %q deleted", c.RelativePath)
		}
	}
}

func (m *FileModule) Rename(args *cmd.FileRenameCmd) {
	if len(args.Files) == 0 {
		m.logger.Error().Msg("No input file")
		return
	}
	m.logger.Info().
		Array("files", extension.StringSlice(args.Files)).
		Str("preset", args.Preset).
		Msgf("Start rename file")

	if args.Preset == "md4" {
		m.RenameByHash(args, args.Preset, "6d6434_")
	}
	if args.Preset == "md5" {
		m.RenameByHash(args, args.Preset, "6d6435_")
	}
	if args.Preset == "sha1" {
		m.RenameByHash(args, args.Preset, "73686131_")
	}
	if args.Preset == "sha256" {
		m.RenameByHash(args, args.Preset, "736861323536_")
	}
	if args.Preset == "sha512" {
		m.RenameByHash(args, args.Preset, "736861353132_")
	}
}

func (m *FileModule) RenameByHash(args *cmd.FileRenameCmd, algo string, prefix string) {
	contents, err := filesystem.List(args.Files, false)
	if err != nil {
		m.logger.Err(err).Msg("Error listing input files")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
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
		m.logger.Info().Str("algo", algo).Int("size", fhResults[0].Size).Msgf("File hashed '%s'", c.RelativePath)
		if err != nil {
			m.logger.Err(err).Msgf("Error computing hash for '%s'", c.RelativePath)
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
		hResults = append(hResults, fhResults...)
	}
	mappings := []*FileRenameMapping{}
	for _, e := range hResults {
		parent := path.Dir(e.Path)
		ext := path.Ext(e.Path)
		targetName := prefix + hex.EncodeToString(e.Hash) + ext
		target := generic.TernaryAssign(parent == ".", targetName, filesystem.Join(parent, targetName))
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
		m.logger.Info().Msgf("Written %d line(s) to '%s'", len(fContents), rollbackFilePath)
	} else {
		m.logger.Err(err).Msgf("Failed to write to '%s'", rollbackFilePath)
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
	}
	for _, e := range mappings {
		if e.Source == e.Target {
			m.logger.Info().Str("src", e.Source).Str("target", e.Target).Msgf("Skip file '%s'", e.Source)
			continue
		}
		if filesystem.IsFileExist(e.Target) {
			m.logger.Error().Str("src", e.Source).Str("target", e.Target).Msgf("Cannot rename '%s'. Target file existed.", e.Source)
			continue
		}
		err := os.Rename(e.Source, e.Target)
		if err != nil {
			m.logger.Err(err).Str("src", e.Source).Str("target", e.Target).Msgf("Cannot rename '%s'", e.Source)
			continue
		}
		m.logger.Info().Str("src", e.Source).Str("target", e.Target).Msgf("Renamed '%s'", e.Source)
	}
}

func (m *FileModule) saveHResults(ctx *db.DbContext, hResults []*core.FileMultiHash, ignore bool, collections []string) (err error) {
	// save Hash
	hashes := make([]*db.Hash, len(hResults))
	for i, res := range hResults {
		hashes[i] = db.NewHash(res, ignore)
	}
	err = ctx.SaveHashes(hashes)
	if err != nil {
		m.logger.Err(err).Msg("Error while saving Hashes.")
		return
	}
	// save Mapping
	sha256s := make([]string, len(hResults))
	for i, res := range hResults {
		sha256s[i] = res.Sha256.HexStr()
	}
	hashes, err = ctx.GetHashesBySha256s(sha256s)
	if err != nil {
		m.logger.Err(err).Msg("Error while reloading Hashes.")
		return
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
		m.logger.Err(err).Msg("Error while saving Mappings.")
		return
	}
	if !slicext.IsEmpty(collections) {
		// save Set
		sets := make([]*db.Set, len(collections))
		for i, name := range collections {
			sets[i] = db.NewSet(name)
		}
		err = ctx.SaveSets(sets)
		if err != nil {
			m.logger.Err(err).Msg("Error while saving Sets.")
			return
		}
		// save SetHash
		sets, err = ctx.GetSetsByNames(collections)
		if err != nil {
			m.logger.Err(err).Msg("Error while reloading Sets.")
			return
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
			m.logger.Err(err).Msg("Error while saving SetHashes.")
			return
		}
	}
	m.logger.Info().Msg("All hashes saved successfully.")
	return
}
