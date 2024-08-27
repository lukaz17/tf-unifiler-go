package main

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/extension"
	"github.com/tforceaio/tf-unifiler-go/extension/generic"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

type HashModule struct {
	logger zerolog.Logger
}

func (m *HashModule) Hash(args *cmd.HashCmd) {
	if args.Create != nil {
		m.CreateHash(args.Create)
	} else {
		m.logger.Error().Msg("Invalid arguments")
	}
}

func (m *HashModule) CreateHash(args *cmd.HashCreateCmd) {
	if len(args.Files) == 0 {
		m.logger.Error().Msg("No input file")
		return
	}
	if len(args.Algorithms) == 0 {
		m.logger.Error().Msg("No hash algorithm")
		return
	}
	m.logger.Info().
		Array("algos", extension.StringSlice(args.Algorithms)).
		Array("files", extension.StringSlice(args.Files)).
		Str("output", args.Output).
		Msgf("Start computing hash")

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
		fhResults, err := hasher.Hash(c.RelativePath, args.Algorithms)
		m.logger.Info().Array("algos", extension.StringSlice(args.Algorithms)).Int("size", fhResults[0].Size).Msgf("File hashed '%s'", c.RelativePath)
		if err != nil {
			m.logger.Err(err).Msgf("Error computing hash for '%s'", c.RelativePath)
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
			return
		}
		hResults = append(hResults, fhResults...)
	}

	for _, a := range args.Algorithms {
		fContents := []string{}
		for _, r := range hResults {
			if a == r.Algorithm {
				line := fmt.Sprintf("%s *%s", hex.EncodeToString(r.Hash), r.Path)
				fContents = append(fContents, line)
			}
		}

		if args.Output == "" {
			m.logger.Warn().Msg("output is not specified, use default name instead")
		}
		output := generic.TernaryAssign(args.Output == "", "checksum", args.Output)
		// substitute file extension. for more information: https://go.dev/play/p/0wZcne8ZC8G
		oPath := fmt.Sprintf("%s.%s", strings.TrimSuffix(output, filepath.Ext(output)), a)
		err := filesystem.WriteLines(oPath, fContents)
		if err == nil {
			m.logger.Info().Msgf("Written %d line(s) to '%s'", len(fContents), oPath)
		} else {
			m.logger.Err(err).Msgf("Failed to write to '%s'", oPath)
			m.logger.Info().Msg("Unexpected error occurred. Exiting...")
		}
	}
}
