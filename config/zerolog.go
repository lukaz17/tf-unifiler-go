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

package config

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
)

func InitZerolog(configDir string) (zerolog.Logger, *os.File) {
	consoleWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.DateTime},
		},
		Level: zerolog.TraceLevel,
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
			consoleLogger := zerolog.New(consoleWriter).With().Timestamp().Logger()
			log.Err(err).Msgf("Cannot create log file: %s", logFile)
			return consoleLogger, nil
		}
	}
	fileWriter, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		consoleLogger := zerolog.New(consoleWriter).With().Timestamp().Logger()
		log.Err(err).Msgf("Cannot create log file: %s", logFile)
		return consoleLogger, nil
	}

	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()
	return logger, nil
}
