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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tforce-io/tf-golib/multiarch"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler-go/db"
)

func TestEmptyCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }() // os.Args is a "global variable", so keep the state from before the test, and restore it after.

	os.Args = []string{"unifiler"}
	main()
	if invokeArgs.Hash != nil {
		t.Errorf("Hash must be nil")
	}
}

func TestHashCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }() // os.Args is a "global variable", so keep the state from before the test, and restore it after.

	os.Args = []string{"unifiler", "hash"}
	main()
	if invokeArgs.Hash == nil {
		t.Errorf("Hash must not be nil")
	}
	if invokeArgs.Hash.Create != nil {
		t.Errorf("Hash.Create must be nil")
	}
}

func TestCreateHashCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }() // os.Args is a "global variable", so keep the state from before the test, and restore it after.

	tests := []struct {
		name   string
		args   []string
		algos  []string
		files  []string
		output string
	}{
		{"CreateHashCommand: 0 algos 1 files", []string{"unifiler", "hash", "create", "-f", "helloworld.txt"}, nil, []string{"helloworld.txt"}, ""},
		{"CreateHashCommand: 1 algos 0 files", []string{"unifiler", "hash", "create", "-a", "sha1", "-o", "/tmp"}, []string{"sha1"}, nil, "/tmp"},
		{"CreateHashCommand: 2 algos 3 files", []string{"unifiler", "hash", "create", "-a", "md5", "sha256", "-f", "helloworld.txt"}, []string{"md5", "sha256"}, []string{"helloworld.txt"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			main()
			if invokeArgs.Hash == nil {
				t.Error("Hash must not be nil.")
			}
			if !reflect.DeepEqual(invokeArgs.Hash.Create.Algorithms, tt.algos) {
				t.Errorf("Incorrect input algorithms. Expected '%s' Actual '%s'", tt.algos, invokeArgs.Hash.Create.Algorithms)
			}
			if !reflect.DeepEqual(invokeArgs.Hash.Create.Files, tt.files) {
				t.Errorf("Incorrect input files. Expected '%s' Actual '%s'", tt.files, invokeArgs.Hash.Create.Files)
			}
			if invokeArgs.Hash.Create.Output != tt.output {
				t.Errorf("Incorrect output files. Expected '%s' Actual '%s'", tt.output, invokeArgs.Hash.Create.Output)
			}
		})
	}
}

func getTestDB(entity, function string) *db.DbContext {
	hasher := sha256.New()
	featSign := fmt.Sprintf("%s/%s/v%d", entity, function, db.SchemaVersion)
	hasher.Write([]byte(featSign))
	hashBuf := hasher.Sum(nil)
	hash := hex.EncodeToString(hashBuf[:])
	fileName := fmt.Sprintf("unifiler_%s.db", hash)
	tmpDir := opx.Ternary(
		multiarch.IsWindows(),
		os.Getenv("TEMP"),
		"/tmp",
	)
	dbFile := filepath.Join(tmpDir, fileName)
	ctx, err := db.Connect(dbFile)
	if err != nil {
		panic(err)
	}
	return ctx
}
