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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/extension"
	"github.com/tforceaio/tf-unifiler-go/extension/generic"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

// HashModule handles user requests related checksum file creation and verification.
type HashModule struct {
	logger zerolog.Logger
}

// Return new HashModule.
func NewHashModule(c *Controller, cmdName string) *HashModule {
	return &HashModule{
		logger: c.CommandLogger("hash", cmdName),
	}
}

// Create checksum file(s) for inputs using 1 or many algorithms.
func (m *HashModule) Create(inputs []string, output string, algorithms []string) error {
	if len(algorithms) == 0 {
		return errors.New("no hash algorithm specified")
	}

	m.logger.Info().
		Array("algos", extension.StringSlice(algorithms)).
		Array("files", extension.StringSlice(inputs)).
		Str("output", output).
		Msgf("Start computing hashes.")

	contents, err := filesystem.List(inputs, true)
	if err != nil {
		return err
	}

	hResults := []*hasher.HashResult{}
	for _, c := range contents {
		if c.IsDir {
			continue
		}
		fhResults, err := hasher.Hash(c.RelativePath, algorithms)
		m.logger.Info().Array("algos", extension.StringSlice(algorithms)).Int("size", fhResults[0].Size).Msgf("File hashed '%s'", c.RelativePath)
		if err != nil {
			return err
		}
		hResults = append(hResults, fhResults...)
	}

	for _, a := range algorithms {
		fContents := []string{}
		for _, r := range hResults {
			if a == r.Algorithm {
				line := fmt.Sprintf("%s *%s", hex.EncodeToString(r.Hash), r.Path)
				fContents = append(fContents, line)
			}
		}

		outputInternal := generic.TernaryAssign(output == "", "checksum", output)
		// substitute file extension. for more information: https://go.dev/play/p/0wZcne8ZC8G
		oPath := fmt.Sprintf("%s.%s", strings.TrimSuffix(outputInternal, filepath.Ext(outputInternal)), a)
		err := filesystem.WriteLines(oPath, fContents)
		if err != nil {
			return err
		}
		m.logger.Info().Msgf("Written %d line(s) to '%s'", len(fContents), oPath)
	}
	return nil
}

func (m *HashModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msgf("Unexpected error has occurred. Program will exit.")
	}
}

// Define Cobra Command for Hash module.
func HashCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "hash",
		Short: "Create and verify checksum files.",
	}

	createCmd := &cobra.Command{
		Use:   "create <output path> <input path> [<input path>...]",
		Short: "Create checksum file(s) using 1 or many hash algorithms.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseHashFlags(cmd)
			m := NewHashModule(c, "create")
			m.logError(m.Create(args[1:], args[0], flags.Algorithms))
		},
	}
	createCmd.Flags().StringSliceP("algo", "a", []string{"sha1"}, "Hash algorithms to use, multiple supported. Valid algorithms: md4, md5, ripemd160, sha1, sha224, sha256, sha384, sha512.")
	rootCmd.AddCommand(createCmd)

	return rootCmd
}

// Struct HashFlags contains all flags used by Hash module.
type HashFlags struct {
	Algorithms []string
}

// Extract all flags from a Cobra Command.
func ParseHashFlags(cmd *cobra.Command) *HashFlags {
	algorithms, _ := cmd.Flags().GetStringSlice("algo")

	return &HashFlags{
		Algorithms: algorithms,
	}
}
