package main

import (
	"os"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog/log"
	"github.com/tforceaio/tf-unifiler-go/cmd"
	"github.com/tforceaio/tf-unifiler-go/crypto/hasher"
	"github.com/tforceaio/tf-unifiler-go/diag"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

var invokeArgs cmd.Args

func main() {
	logFile := diag.InitZerolog()
	if logFile != nil {
		defer logFile.Close()
	}
	filesystem.SetLogger(diag.GetModuleLogger("filesystem"))
	hasher.SetLogger(diag.GetModuleLogger("crypto/hasher"))

	invokeArgs = cmd.Args{}
	arg.MustParse(&invokeArgs)
	pwd, _ := os.Getwd()
	log.Info().Msgf("Current directory %s", pwd)

	if invokeArgs.Hash != nil {
		m := HashModule{
			logger: log.Logger.With().Str("module", "hash").Logger(),
		}
		m.Hash(invokeArgs.Hash)
	}
}
