// Copyright (C) 2025 T-Force I/O
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
	"os"

	"github.com/rs/zerolog"
)

type Controller struct {
	Root   *RootConfig
	Logger zerolog.Logger

	logFile *os.File
}

func Init() *Controller {
	config, err := InitKoanf()
	logger, logFile := InitZerolog(config.ConfigDir)
	if err != nil {
		logger.Err(err).Msg("error initializing config")
	}
	return &Controller{
		Root:   config,
		Logger: logger,

		logFile: logFile,
	}
}

func (c *Controller) Close() {
	if c.logFile != nil {
		c.logFile.Close()
		c.logFile = nil
	}
}

func (c *Controller) CommandLogger(module, command string) zerolog.Logger {
	return c.Logger.With().Str("module", module).Str("command", command).Logger()
}

func (c *Controller) ModuleLogger(module string) zerolog.Logger {
	return c.Logger.With().Str("module", module).Logger()
}
