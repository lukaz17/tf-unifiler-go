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

	"github.com/rs/zerolog"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/config"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/extension"
	"github.com/tforceaio/tf-unifiler-go/extension/generic"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
	"github.com/tforceaio/tf-unifiler-go/parser"
)

type MirrorFileMapping struct {
	Source string `json:"s,omitempty"`
	Hash   string `json:"h,omitempty"`
}

type MirrorModule struct {
	logger zerolog.Logger
}

func NewMirrorModule(cfg *config.Controller) *MirrorModule {
	return &MirrorModule{
		logger: cfg.ModuleLogger("Mirror"),
	}
}

func (m *MirrorModule) Mirror(args *cmd.MirrorCmd) {
	if args.Export != nil {
		m.export(args.Export)
	} else if args.Import != nil {
		m.import2(args.Import)
	} else if args.Scan != nil {
		m.scan(args.Scan)
	} else {
		m.logger.Error().Msg("Invalid arguments")
	}
}

func (m *MirrorModule) export(args *cmd.MirrorExportCmd) {
	if args.Cache == "" {
		m.logger.Error().Msg("Cache not set")
		return
	} else if !filesystem.IsDirectoryExist(args.Cache) {
		m.logger.Error().Str("path", args.Cache).Msg("Cache path not found")
		return
	}
	if args.Checksum == "" {
		m.logger.Error().Msg("Checksum file not set")
		return
	} else if !filesystem.IsFileExist(args.Checksum) {
		m.logger.Error().Str("path", args.Cache).Msg("Checksum file not found")
		return
	}
	m.logger.Info().
		Str("cache", args.Cache).
		Str("checksum", args.Checksum).
		Str("root", args.TargetRoot).
		Msgf("Start exporting files structure")

	checksumReader, err := os.OpenFile(args.Checksum, os.O_RDONLY, 0664)
	if err != nil {
		m.logger.Err(err).Str("path", args.Cache).Msg("Cannot read checksum file")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}
	items, err := parser.ParseSha256(checksumReader)
	if err != nil {
		m.logger.Err(err).Str("path", args.Cache).Msg("Invalid checksum file")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}
	defer checksumReader.Close()

	missingItems := []string{}
	for _, l := range items {
		cachePath := path.Join(args.Cache, l.Hash)
		if !filesystem.IsFileExist(cachePath) {
			missingItems = append(missingItems, l.Hash)
		}
	}
	if len(missingItems) > 0 {
		m.logger.Error().Array("hashes", extension.StringSlice(missingItems)).Msg("Missing cache items")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}

	if args.TargetRoot == "" {
		m.logger.Warn().Msg("target root is not specified, it will be derived from checkfile path instead")
	} else {
		if filesystem.IsFileExist(args.TargetRoot) {
			m.logger.Error().Msg("A file with same with target root existed")
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
	}
	targetRoot := args.TargetRoot
	if targetRoot == "" {
		checksumPath, _ := filesystem.GetAbsPath(args.Checksum)
		targetRoot, _ = path.Split(checksumPath)
	} else {
		targetRoot, _ = filesystem.GetAbsPath(args.TargetRoot)
	}
	for _, l := range items {
		cachePath := path.Join(args.Cache, l.Hash)
		targetPath := generic.TernaryAssign(filesystem.IsAbsPath(l.Path), l.Path, path.Join(targetRoot, l.Path))
		err := filesystem.CreateHardlink(cachePath, targetPath)
		if err != nil {
			m.logger.Err(err).Str("src", cachePath).Str("target", targetPath).Msg("Error creating hardlink")
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		} else {
			m.logger.Info().Str("src", cachePath).Str("target", targetPath).Msgf("Exported file '%s'", l.Hash)
		}
	}
}

func (m *MirrorModule) import2(args *cmd.MirrorImportCmd) {
	if args.Cache == "" {
		m.logger.Error().Msg("Cache not set")
		return
	}
	if args.Checksum == "" {
		m.logger.Error().Msg("Checksum file not set")
		return
	} else if !filesystem.IsFileExist(args.Checksum) {
		m.logger.Error().Str("path", args.Cache).Msg("Checksum file not found")
		return
	}
	m.logger.Info().
		Str("cache", args.Cache).
		Str("checksum", args.Checksum).
		Str("root", args.TargetRoot).
		Msgf("Start importing files structure")

	checksumReader, err := os.OpenFile(args.Checksum, os.O_RDONLY, 0664)
	if err != nil {
		m.logger.Err(err).Str("path", args.Cache).Msg("Cannot read checksum file")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}
	items, err := parser.ParseSha256(checksumReader)
	if err != nil {
		m.logger.Err(err).Str("path", args.Cache).Msg("Invalid checksum file")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}
	defer checksumReader.Close()

	if args.TargetRoot == "" {
		m.logger.Warn().Msg("target root is not specified, it will be derived from checkfile path instead")
	} else {
		if filesystem.IsFileExist(args.TargetRoot) {
			m.logger.Error().Msg("A file with same with target root existed")
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
	}
	targetRoot := args.TargetRoot
	if targetRoot == "" {
		checksumPath, _ := filesystem.GetAbsPath(args.Checksum)
		targetRoot, _ = path.Split(checksumPath)
	} else {
		targetRoot, _ = filesystem.GetAbsPath(args.TargetRoot)
	}

	missingItems := []string{}
	for _, l := range items {
		sourcePath := generic.TernaryAssign(filesystem.IsAbsPath(l.Path), l.Path, path.Join(targetRoot, l.Path))
		if !filesystem.IsFileExist(sourcePath) {
			missingItems = append(missingItems, l.Path)
		}
	}
	if len(missingItems) > 0 {
		m.logger.Error().Array("files", extension.StringSlice(missingItems)).Msg("Missing source files")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}

	for _, l := range items {
		sourcePath := generic.TernaryAssign(filesystem.IsAbsPath(l.Path), l.Path, path.Join(targetRoot, l.Path))
		cachePath := path.Join(args.Cache, l.Hash)
		err := filesystem.CreateHardlink(sourcePath, cachePath)
		if err != nil {
			m.logger.Err(err).Str("src", sourcePath).Str("target", cachePath).Msg("Error creating hardlink")
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		} else {
			m.logger.Info().Str("src", sourcePath).Str("target", cachePath).Msgf("Exported file '%s'", l.Hash)
		}
	}
}

func (m *MirrorModule) scan(args *cmd.MirrorScanCmd) {
	if args.Cache == "" {
		m.logger.Error().Msg("Cache not set")
		return
	} else if !filesystem.IsDirectoryExist(args.Cache) {
		m.logger.Error().Str("path", args.Cache).Msg("Cache path not found")
		return
	}
	if len(args.Files) == 0 {
		m.logger.Error().Msg("No input file")
		return
	}
	m.logger.Info().
		Str("cache", args.Cache).
		Array("files", extension.StringSlice(args.Files)).
		Msgf("Start scanning files")

	contents, err := filesystem.List(args.Files, true)
	if err != nil {
		m.logger.Err(err).Msg("Error listing input files")
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		return
	}

	hResults := []*hasher.HashResult{}
	for _, c := range contents {
		if c.IsDir {
			continue
		}
		fhResult, err := hasher.HashSha256(c.RelativePath)
		if err != nil {
			m.logger.Err(err).Msgf("Error computing hash for '%s'", c.RelativePath)
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
		m.logger.Info().Str("algo", "sha256").Int("size", fhResult.Size).Msgf("File hashed '%s'", c.RelativePath)
		fhResult.Path = c.AbsolutePath
		hResults = append(hResults, fhResult)
	}

	mappings := []*MirrorFileMapping{}
	for _, e := range hResults {
		mapping := &MirrorFileMapping{
			Source: e.Path,
			Hash:   hex.EncodeToString(e.Hash),
		}
		mappings = append(mappings, mapping)
	}
	currentTimestamp := time.Now().UnixMilli()
	rollbackFilePath := filesystem.Join(args.Cache, "unifiler-mirror-"+strconv.FormatInt(currentTimestamp, 10)+".json")
	fContent, _ := json.Marshal(mappings)
	fContents := []string{string(fContent)}
	err = filesystem.WriteLines(rollbackFilePath, fContents)
	if err == nil {
		m.logger.Info().Msgf("Written %d line(s) to '%s'", len(mappings), rollbackFilePath)
	} else {
		m.logger.Err(err).Msgf("Failed to write to '%s'", rollbackFilePath)
		m.logger.Info().Msg("Unexpected error occurred. Exiting...")
	}
	for _, r := range hResults {
		name := hex.EncodeToString(r.Hash)
		cachePath := path.Join(args.Cache, name)
		if filesystem.IsFileExist(cachePath) {
			m.logger.Info().Str("src", r.Path).Str("cache", cachePath).Msg("File is already cached")
		} else {
			err := filesystem.CreateHardlink(r.Path, cachePath)
			if err != nil {
				m.logger.Err(err).Str("src", r.Path).Str("target", cachePath).Msg("Error creating hardlink")
				m.logger.Info().Msg("Unexpected error occurred. Exiting...")
				return
			}
			m.logger.Info().Str("src", r.Path).Str("target", cachePath).Msgf("Created cache file")
		}
	}
}
