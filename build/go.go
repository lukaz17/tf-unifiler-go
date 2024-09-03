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
