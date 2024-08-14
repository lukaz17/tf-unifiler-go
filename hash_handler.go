package main

import (
	"fmt"

	"github.com/tforceaio/tf-unifiler-go/cmd"
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
	contents, err := filesystem.List(args.Files, true)
	if err != nil {
		fmt.Printf("CreateHash: Error listing input files. %v", err)
	}
	for _, e := range contents {
		println(e.RelativePath)
	}
}
