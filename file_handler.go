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

package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/rs/zerolog"
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

type FileRenameMapping struct {
	Source string `json:"s,omitempty"`
	Target string `json:"t,omitempty"`
}

func (m *FileModule) File(args *cmd.FileCmd) {
	if args.Delete != nil {
		m.Delete(args.Delete)
	} else if args.Rename != nil {
		m.Rename(args.Rename)
	} else {
		m.logger.Error().Msg("Invalid arguments")
	}
}

func (m *FileModule) Delete(args *cmd.FileDeleteCmd) {
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

	hResults := []*db.Hash{}
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
		hResults = append(hResults, db.NewHash(fileMultiHash, true))
	}

	dbFile := filesystem.Join(args.Workspace, "metadata.db")
	ctx, err := db.Connect(dbFile)
	if err != nil {
		m.logger.Err(err).Msg("Error while opening metadata database.")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}
	err = ctx.SaveHashes(hResults)
	if err != nil {
		m.logger.Err(err).Msg("Error while saving ignored files.")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}
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
