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

package parser

import (
	"fmt"
	"io"
	"regexp"

	"github.com/tforceaio/tf-unifiler-go/parser/checksum"
)

func ParseSha256(r io.Reader) ([]*checksum.ChecksumItem, error) {
	parser := checksum.NewParser(r)
	items, err := parser.Parse()
	if err != nil {
		return items, err
	}
	hashCheckRegex := regexp.MustCompile("^[0-9A-Fa-f]{64}$")
	for _, l := range items {
		if !hashCheckRegex.MatchString(l.Hash) {
			return []*checksum.ChecksumItem{}, fmt.Errorf("invalid SHA-256 hash '%s'", l.Hash)
		}
	}
	return items, err
}
