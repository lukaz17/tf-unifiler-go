package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

var GOBIN, _ = filepath.Abs(".bin")

func executablePath(name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(GOBIN, name)
}

func main() {
	log.SetFlags(log.Lshortfile)

	if !filesystem.IsFileExist(filepath.Join("build", "ci.go")) {
		log.Fatal("this script must be run from the root of the repository")
	}

	compile(os.Args[1:])
}

func compile(cmdline []string) {
	var (
		arch       = flag.String("arch", "", "Architecture to cross build for")
		cc         = flag.String("cc", "", "C compiler to cross build with")
		staticlink = flag.Bool("static", false, "Create statically-linked executable")
	)
	flag.CommandLine.Parse(cmdline)
	env := Env()

	// Configure the toolchain.
	tc := GoToolchain{GOARCH: *arch, CC: *cc}
	// Disable CLI markdown doc generation in release builds.
	buildTags := []string{"urfave_cli_no_docs"}

	// Enable linking the CKZG library since we can make it work with additional flags.
	// Assume all version of host OS is not Ubuntu trusty tahr
	buildTags = append(buildTags, "ckzg")

	// Configure the build.
	gobuild := tc.Go("build", buildFlags(env, *staticlink, buildTags)...)

	// arm64 CI builders are memory-constrained and can't handle concurrent builds,
	// better disable it. This check isn't the best, it should probably
	// check for something in env instead.
	if runtime.GOARCH == "arm64" {
		gobuild.Args = append(gobuild.Args, "-p", "1")
	}
	// We use -trimpath to avoid leaking local paths into the built executables.
	gobuild.Args = append(gobuild.Args, "-trimpath")

	// Show packages during build.
	gobuild.Args = append(gobuild.Args, "-v")

	// Now we choose what we're even building.
	// Default: collect all 'main' packages in cmd/ and build those.
	packages := []struct {
		class string
		name  string
	}{
		{"./.", "unifiler"},
	}

	// Do the build!
	for _, pkg := range packages {
		args := make([]string, len(gobuild.Args))
		copy(args, gobuild.Args)
		args = append(args, "-o", executablePath(pkg.name))
		args = append(args, pkg.class)
		MustRun(&exec.Cmd{Path: gobuild.Path, Args: args, Env: gobuild.Env})
	}
}

// buildFlags returns the go tool flags for building.
func buildFlags(env Environment, staticLinking bool, buildTags []string) (flags []string) {
	var ld []string
	// See https://github.com/golang/go/issues/33772#issuecomment-528176001
	// We need to set --buildid to the linker here, and also pass --build-id to the
	// cgo-linker further down.
	ld = append(ld, "--buildid=none")

	if env.Commit != "" {
		ld = append(ld, "-X", "main.gitCommit="+fmt.Sprintf("%.8s", env.Commit))
		ld = append(ld, "-X", "main.gitDate="+env.Date)
		ld = append(ld, "-X", "main.gitBranch="+env.Branch)
	}
	// Omit debug information to reduce file size
	// See https://go.dev/doc/gdb#Introduction
	if env.Branch == "master" {
		ld = append(ld, "-w")
	}
	// Strip DWARF on darwin. This used to be required for certain things,
	// and there is no downside to this, so we just keep doing it.
	if runtime.GOOS == "darwin" {
		ld = append(ld, "-s")
	}
	if runtime.GOOS == "linux" {
		// Enforce the stacksize to 8M, which is the case on most platforms apart from
		// alpine Linux.
		// See https://sourceware.org/binutils/docs-2.23.1/ld/Options.html#Options
		// regarding the options --build-id=none and --strip-all. It is needed for
		// reproducible builds; removing references to temporary files in C-land, and
		// making build-id reproducably absent.
		extld := []string{"-Wl,-z,stack-size=0x800000,--build-id=none,--strip-all"}
		if staticLinking {
			extld = append(extld, "-static")
			// Under static linking, use of certain glibc features must be
			// disabled to avoid shared library dependencies.
			buildTags = append(buildTags, "osusergo", "netgo")
		}
		ld = append(ld, "-extldflags", "'"+strings.Join(extld, " ")+"'")
	}
	if len(ld) > 0 {
		flags = append(flags, "-ldflags", strings.Join(ld, " "))
	}
	if len(buildTags) > 0 {
		flags = append(flags, "-tags", strings.Join(buildTags, ","))
	}
	return flags
}
