package main

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/extension/generic"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

func Hash(args *cmd.HashCmd) {
	if args.Create != nil {
		CreateHash(args.Create)
	} else {
		println("Hash: Invalid arguments")
	}
}

func CreateHash(args *cmd.HashCreateCmd) {
	if len(args.Files) == 0 {
		println("CreateHash: No input file")
		return
	}
	if len(args.Algorithms) == 0 {
		println("CreateHash: No hash algorithm")
		return
	}
	contents, err := filesystem.List(args.Files, true)
	if err != nil {
		fmt.Printf("CreateHash: Error listing input files. %v", err)
	}

	hResults := []*hasher.HashResult{}
	for _, c := range contents {
		if c.IsDir {
			continue
		}
		fhResults, err := hasher.Hash(c.RelativePath, args.Algorithms)
		if err != nil {
			fmt.Printf("CreateHash: Failed to hash '%s'. %v", c.RelativePath, err)
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

		output := generic.TernaryAssign(args.Output == "", "checksum", args.Output)
		// substitute file extension. for more information: https://go.dev/play/p/0wZcne8ZC8G
		oPath := fmt.Sprintf("%s.%s", strings.TrimSuffix(output, filepath.Ext(output)), a)
		err := filesystem.WriteLines(oPath, fContents)
		if err != nil {
			fmt.Printf("Failed to write to '%s'. %v", oPath, err)
		}
	}
}
