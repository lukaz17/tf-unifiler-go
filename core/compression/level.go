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

package compression

import (
	"errors"
	"fmt"
)

type Level int

const (
	None Level = iota
	Fast
	Normal
	High
	Ultra
)

var LevelNames = map[string]Level{
	"none":   None,
	"fast":   Fast,
	"normal": Normal,
	"high":   High,
	"ultra":  Ultra,
}

func ParseLevel(name string) (Level, error) {
	if level, ok := LevelNames[name]; ok {
		return level, nil
	}

	var num int
	if _, err := fmt.Sscanf(name, "%d", &num); err == nil && num >= int(None) && num <= int(Ultra) {
		return Level(num), nil
	}

	return 0, errors.New("invalid level: " + name)
}
