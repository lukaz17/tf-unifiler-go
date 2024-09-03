package diag

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitZerolog() *os.File {
	exec, _ := os.Executable()
	exPath := filepath.Dir(exec)
	date := time.Now().UTC().Format("20060102")
	logFile := path.Join(exPath, fmt.Sprintf("unifiler-%s.log", date))

	consoleWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{zerolog.ConsoleWriter{Out: os.Stdout, NoColor: false, TimeFormat: time.DateTime}},
		Level:  zerolog.InfoLevel,
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
