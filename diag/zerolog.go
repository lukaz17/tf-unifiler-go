package diag

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

func InitZerolog(configDir string) *os.File {
	consoleWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false, TimeFormat: time.DateTime}},
		Level:  zerolog.InfoLevel,
	}

	logFile := ""
	if configDir != "" {
		date := time.Now().UTC().Format("20060102")
		logFile = path.Join(configDir, "logs", fmt.Sprintf("unifiler-%s.log", date))
	}
	logDir := path.Join(configDir, "logs")
	if !filesystem.IsExist(logDir) {
		err := filesystem.CreateDirectoryRecursive(logDir)
		if err != nil {
			log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
			log.Err(err).Msgf("Cannot create log file: %s", logFile)
			return nil
		}
	}
	fileWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
		log.Err(err).Msgf("Cannot create log file: %s", logFile)
		return nil
	}

	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Logger()
	return fileWriter
}

func GetModuleLogger(name string) zerolog.Logger {
	return log.Logger.With().Str("module", name).Logger()
}
