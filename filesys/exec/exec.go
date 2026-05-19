// Copyright (C) 2025 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

package exec

import (
	"os"
	"os/exec"
)

type CommandArgs interface {
	Compile() []string
}

// Run an app and wait for it to finish. Passthrough will redirect the app stdin, stdout, and stderr to current process.
func Run(app string, arg CommandArgs, passthrough bool) (string, error) {
	cmd := exec.Command(app, arg.Compile()...)
	if passthrough {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		return "", err
	}

	stdout, err := cmd.Output()
	return string(stdout), err
}
