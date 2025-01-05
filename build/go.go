// Copyright (C) 2024 T-Force I/O
// This file is part of TF Unifiler Build library.
//
// TF Unifiler Build library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TF Unifiler Build library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with TF Unifiler Build library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is a modified version of The go-ethereum library, Copyright 2016 The go-ethereum Authors

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type GoToolchain struct {
	Root string // GOROOT

	// Cross-compilation variables. These are set when running the go tool.
	GOARCH string
	GOOS   string
	CC     string
}

// Go creates an invocation of the go command.
func (g *GoToolchain) Go(command string, args ...string) *exec.Cmd {
	tool := g.goTool(command, args...)

	// Configure environment for cross build.
	if g.GOARCH != "" && g.GOARCH != runtime.GOARCH {
		tool.Env = append(tool.Env, "CGO_ENABLED=1")
		tool.Env = append(tool.Env, "GOARCH="+g.GOARCH)
	}
	if g.GOOS != "" && g.GOOS != runtime.GOOS {
		tool.Env = append(tool.Env, "GOOS="+g.GOOS)
	}
	// Configure C compiler.
	if g.CC != "" {
		tool.Env = append(tool.Env, "CC="+g.CC)
	} else if os.Getenv("CC") != "" {
		tool.Env = append(tool.Env, "CC="+os.Getenv("CC"))
	}
	// CKZG by default is not portable, append the necessary build flags to make
	// it not rely on modern CPU instructions and enable linking against.
	tool.Env = append(tool.Env, "CGO_CFLAGS=-O2 -g -D__BLST_PORTABLE__")

	return tool
}

func (g *GoToolchain) goTool(command string, args ...string) *exec.Cmd {
	if g.Root == "" {
		g.Root = runtime.GOROOT()
	}
	tool := exec.Command(filepath.Join(g.Root, "bin", "go"), command) // nolint: gosec
	tool.Args = append(tool.Args, args...)
	tool.Env = append(tool.Env, "GOROOT="+g.Root)

	// Forward environment variables to the tool, but skip compiler target settings.
	// TODO: what about GOARM?
	skip := map[string]struct{}{"GOROOT": {}, "GOARCH": {}, "GOOS": {}, "GOBIN": {}, "CC": {}}
	for _, e := range os.Environ() {
		if i := strings.IndexByte(e, '='); i >= 0 {
			if _, ok := skip[e[:i]]; ok {
				continue
			}
		}
		tool.Env = append(tool.Env, e)
	}
	return tool
}
