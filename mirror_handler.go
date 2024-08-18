package main

import (
	"encoding/hex"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/extension"
	"github.com/tforceaio/tf-unifiler-go/extension/generic"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
	"github.com/tforceaio/tf-unifiler-go/parser"
)

type MirrorModule struct {
	logger zerolog.Logger
}

func (m *MirrorModule) Mirror(args *cmd.MirrorCmd) {
	if args.Export != nil {
		m.MirrorExport(args.Export)
	} else if args.Scan != nil {
		m.MirrorScan(args.Scan)
	} else {
		m.logger.Error().Msg("Invalid arguments")
	}
}

func (m *MirrorModule) MirrorExport(args *cmd.MirrorExportCmd) {
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
		} else {
			m.logger.Info().Str("src", cachePath).Str("target", targetPath).Msgf("Exported file '%s'", l.Hash)
		}
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
				m.logger.Err(err).Str("src", r.Path).Str("target", cachePath).Msg("Error creating hardlink")
				m.logger.Info().Msg("Unexpected error occurred. Exiting...")
				return
			}
			m.logger.Info().Str("src", r.Path).Str("target", cachePath).Msgf("Created cache file")
		}
	}
}
