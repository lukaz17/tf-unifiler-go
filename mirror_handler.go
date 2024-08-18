package main

import (
	"encoding/hex"
	"path"

	"github.com/rs/zerolog"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/extension"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

type MirrorModule struct {
	logger zerolog.Logger
}

func (m *MirrorModule) Mirror(args *cmd.MirrorCmd) {
	if args.Scan != nil {
		m.MirrorScan(args.Scan)
	} else {
		m.logger.Error().Msg("Invalid arguments")
	}
}

func (m *MirrorModule) MirrorScan(args *cmd.MirrorScanCmd) {
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
		fhResult.Path = c.AbsolutePath
		hResults = append(hResults, fhResult)
	}

	for _, r := range hResults {
		name := hex.EncodeToString(r.Hash)
		cachePath := path.Join(args.Cache, name)
		if filesystem.IsFileExist(cachePath) {
			m.logger.Info().Str("src", r.Path).Str("cache", cachePath).Msg("File is already cached")
		} else {
			err := filesystem.CreateHardlink(r.Path, cachePath)
			if err != nil {
				m.logger.Err(err).Str("path", r.Path).Msg("Error creating hardlink")
				m.logger.Info().Msg("Unexpected error occurred. Exiting...")
				return
			}
			m.logger.Info().Str("src", r.Path).Str("target", cachePath).Msgf("Created cache file")
		}
	}
}
