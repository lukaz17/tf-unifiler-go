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

//go:build windows

package filesys

import (
	"time"

	"golang.org/x/sys/windows"
)

// Set Archive attribute of a file or folder.
func SetArchiveAttribute(path string, enable bool) error {
	return setFileAttribute(path, windows.FILE_ATTRIBUTE_ARCHIVE, enable)
}

// Set Created time of a file or folder.
func SetCreatedTime(path string, ctime time.Time) error {
	ft := windows.NsecToFiletime(ctime.UnixNano())
	handle, err := windows.CreateFile(windows.StringToUTF16Ptr(path),
		windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	var stime, atime, mtime windows.Filetime
	err = windows.GetFileTime(handle, &stime, &atime, &mtime)
	if err != nil {
		return err
	}
	return windows.SetFileTime(handle, &ft, &atime, &mtime)
}

// Set Hidden attribute of a file or folder.
func SetHiddenAttribute(path string, enable bool) error {
	return setFileAttribute(path, windows.FILE_ATTRIBUTE_HIDDEN, enable)
}

// Set Read Only attribute of a file or folder.
func SetReadOnlyAttribute(path string, enable bool) error {
	return setFileAttribute(path, windows.FILE_ATTRIBUTE_READONLY, enable)
}

// Set System attribute of a file or folder.
func SetSystemAttribute(path string, enable bool) error {
	return setFileAttribute(path, windows.FILE_ATTRIBUTE_SYSTEM, enable)
}

// Set attribute value for file or folder on Windows.
func setFileAttribute(path string, flag uint32, enable bool) error {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return err
	}
	if enable {
		attrs |= flag
	} else {
		attrs &^= flag
	}
	return windows.SetFileAttributes(ptr, attrs)
}
