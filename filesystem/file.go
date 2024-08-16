package filesystem

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/tforceaio/tf-unifiler-go/extension/generic"
)

type FsEntry struct {
	AbsolutePath string
	RelativePath string
	Name         string
	IsDir        bool
}

func CreateEntry(fPath string) (*FsEntry, error) {
	absolutePath, err := filepath.Abs(fPath)
	if err != nil {
		return nil, err
	}
	fileInfo, err := os.Lstat(fPath)
	if err != nil {
		return nil, err
	}
	entry := &FsEntry{
		AbsolutePath: absolutePath,
		RelativePath: fPath,
		Name:         fileInfo.Name(),
		IsDir:        fileInfo.IsDir(),
	}
	return entry, nil
}

func IsExist(fPath string) bool {
	_, err := os.Stat(fPath)
	return !os.IsNotExist(err)
}

func IsFile(fPath string) (bool, error) {
	fileInfo, err := os.Lstat(fPath)
	if err != nil {
		return false, err
	}
	return !fileInfo.IsDir(), nil
}

func IsFileUnsafe(fPath string) bool {
	isFile, err := IsFile(fPath)
	if err != nil {
		panic(err)
	}
	return isFile
}

func IsFileExist(fPath string) bool {
	fileInfo, err := os.Stat(fPath)
	if os.IsNotExist(err) {
		return false
	}
	return !fileInfo.IsDir()
}

func GetPaths(entries []*FsEntry) []string {
	fPaths := make([]string, len(entries))
	for i, e := range entries {
		fPaths[i] = e.RelativePath
	}
	return fPaths
}

func GetAbsPaths(entries []*FsEntry) []string {
	fPaths := make([]string, len(entries))
	for i, e := range entries {
		fPaths[i] = e.AbsolutePath
	}
	return fPaths
}

func List(fPaths []string, recursive bool) ([]*FsEntry, error) {
	contents := make([]*FsEntry, len(fPaths))
	for i, p := range fPaths {
		entry, err := CreateEntry(p)
		if err != nil {
			return []*FsEntry{}, err
		}
		contents[i] = entry
	}
	maxDepth := generic.TernaryAssign(recursive, -1, 1)
	if recursive {
		var err error
		contents, err = listEntries(contents, maxDepth, 0)
		if err != nil {
			return []*FsEntry{}, err
		}
	}
	return contents, nil
}

func WriteLines(fPath string, lines []string) error {
	f, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(f)
	for _, line := range lines {
		writer.WriteString(line)
		writer.WriteString("\n")
	}
	writer.Flush()

	return nil
}
