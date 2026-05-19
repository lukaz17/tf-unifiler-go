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

package engine

import (
	"github.com/tforceaio/tf-unifiler/core/compression"
)

func to7zCompressLevel(level compression.Level) int {
	switch level {
	case compression.None:
		return 0
	case compression.Fast:
		return 3
	case compression.Normal:
		return 5
	case compression.High:
		return 7
	case compression.Ultra:
		return 9
	default:
		return 5
	}
}

func toRarCompressLevel(level compression.Level) int {
	switch level {
	case compression.None:
		return 0
	case compression.Fast:
		return 2
	case compression.Normal:
		return 3
	case compression.High:
		return 4
	case compression.Ultra:
		return 5
	default:
		return 3
	}
}
