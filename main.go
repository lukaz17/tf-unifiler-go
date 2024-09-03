package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog/log"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/diag"
	"github.com/tforceaio/tf-unifiler-go/extension/generic"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
	"github.com/tforceaio/tf-unifiler-go/filesystem/exec"
)

var invokeArgs cmd.Args
var majorVersion = 0
var minorVersion = 2
var patchVersion = 1
var gitCommit, gitDate, gitBranch string

func version() string {
	originDate := time.Date(2024, time.August, 13, 0, 0, 0, 0, time.UTC)
	gitDate2, _ := time.Parse("20060102", gitDate)
	buildDate := generic.TernaryAssign(gitDate == "", time.Now().UTC(), gitDate2)
	duration := buildDate.Sub(originDate)
	minor := minorVersion
	patch := strconv.Itoa(patchVersion)
	if gitBranch == "master" {
		// do nothing
	} else if gitBranch == "release" {
		minor += 1
		patch = patch + "-rc"
	} else if strings.Contains(gitBranch, "feat/") {
		minor += 1
		patch = patch + "-dev"
	} else {
		patch = strconv.Itoa(patchVersion+1) + "-dev"
	}
	if gitCommit != "" {
		return fmt.Sprintf("%d.%d.%s.%d-%s", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000), gitCommit)
	}
	return fmt.Sprintf("%d.%d.%s.%d", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000))
}

func main() {
	logFile := diag.InitZerolog()
	if logFile != nil {
		defer logFile.Close()
	}
	filesystem.SetLogger(diag.GetModuleLogger("filesystem"))
	exec.SetLogger(diag.GetModuleLogger("exec"))

	invokeArgs = cmd.Args{}
	arg.MustParse(&invokeArgs)
	pwd, _ := os.Getwd()
	pwd, _ = filesystem.GetAbsPath(pwd)
	exec, _ := os.Executable()
	exec, _ = filesystem.GetAbsPath(exec)

	log.Info().Msgf("TF UNIFILER v%s", version())
	log.Info().Msgf("Working directory %s", pwd)
	log.Info().Msgf("Executable file %s", exec)

	if invokeArgs.Hash != nil {
		m := HashModule{
			logger: diag.GetModuleLogger("hash"),
		}
		m.Hash(invokeArgs.Hash)
	}
	if invokeArgs.Mirror != nil {
		m := MirrorModule{
			logger: diag.GetModuleLogger("mirror"),
		}
		m.Mirror(invokeArgs.Mirror)
	}
	if invokeArgs.Video != nil {
		m := VideoModule{
			logger: diag.GetModuleLogger("video"),
		}
		m.Video(invokeArgs.Video)
	}
}
