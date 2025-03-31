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

package engine

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tforceaio/tf-unifiler-go/extension/generic"
)

// Execute the program.
func Execute() {
	gitDate2, _ := time.Parse("20060102", gitDate)
	buildDate := generic.TernaryAssign(gitDate == "", time.Now().UTC(), gitDate2)

	rootCmd := &cobra.Command{
		Use: "unifiler",
		Long: fmt.Sprintf(
			`TF UNIFILER v%s.
Copyright (C) %d T-Force I/O.
Licensed under GPL-3.0 license. See COPYING file along with this program for more details.`,
			version(),
			buildDate.Year()),
		Short:   "Cross platform file managements command line utility.",
		Version: version(),
	}
	rootCmd.AddCommand(ChecksumCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
