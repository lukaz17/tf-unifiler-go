package main

import (
	"os"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog/log"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/diag"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
	"github.com/tforceaio/tf-unifiler-go/filesystem/exec"
)

var invokeArgs cmd.Args
var version = "v0.1.0"

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
	log.Info().Msgf("TF UNIFILER %s", version)
	log.Info().Msgf("Current directory %s", pwd)

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
