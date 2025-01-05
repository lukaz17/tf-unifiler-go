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
