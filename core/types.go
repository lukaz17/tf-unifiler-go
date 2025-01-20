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

package core

import "github.com/tforce-io/tf-golib/stdx"

type FileMultiHash struct {
	Md5      stdx.Bytes
	Sha1     stdx.Bytes
	Sha256   stdx.Bytes
	Sha512   stdx.Bytes
	Size     uint32
	FileName string
}
