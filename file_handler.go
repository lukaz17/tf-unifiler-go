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
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
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
	if args.Rename != nil {
		m.Rename(args.Rename)
	} else {
		m.logger.Error().Msg("Invalid arguments")
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
