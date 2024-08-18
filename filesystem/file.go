package filesystem

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/tforceaio/tf-unifiler-go/extension/generic"
)

type FsEntry struct {
	AbsolutePath string
	RelativePath string
	Name         string
	IsDir        bool
}

type FsEntries []*FsEntry

func (entries FsEntries) GetPaths() []string {
	fPaths := make([]string, len(entries))
	for i, e := range entries {
		fPaths[i] = e.RelativePath
	}
	return fPaths
}

func (entries FsEntries) GetAbsPaths() []string {
	fPaths := make([]string, len(entries))
	for i, e := range entries {
		fPaths[i] = e.AbsolutePath
	}
	return fPaths
}

func CreateEntry(fPath string) (*FsEntry, error) {
	absolutePath, err := GetAbsPath(fPath)
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

func CreateHardlink(sPath, tPath string) error {
	err := os.Link(sPath, tPath)
	if err == nil {
		logger.Debug().Str("src", sPath).Str("target", tPath).Msgf("Create link for '%s'", sPath)
	}
	return err
}

func GetAbsPath(fPath string) (string, error) {
	absolutePath, err := filepath.Abs(fPath)
	if err == nil {
		absolutePath = strings.ReplaceAll(absolutePath, "\\", "/") // enfore linux path style for clarity
	}
	return absolutePath, err
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

func List(fPaths []string, recursive bool) (FsEntries, error) {
	contents := make([]*FsEntry, len(fPaths))
	for i, p := range fPaths {
		entry, err := CreateEntry(p)
		if err != nil {
			return FsEntries{}, err
		}
		contents[i] = entry
	}
	maxDepth := generic.TernaryAssign(recursive, -1, 1)
	if recursive {
		var err error
		contents, err = listEntries(contents, maxDepth, 0)
		if err != nil {
			return FsEntries{}, err
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
