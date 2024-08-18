package filesystem

import (
	"os"
	"path"
)

func CreateDirectory(dPath string) error {
	return os.Mkdir(dPath, 0755)
}

func IsDirectory(dPath string) (bool, error) {
	fileInfo, err := os.Lstat(dPath)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func IsDirectoryUnsafe(dPath string) bool {
	isDir, err := IsDirectory(dPath)
	if err != nil {
		panic(err)
	}
	return isDir
}

func IsDirectoryExist(fPath string) bool {
	fileInfo, err := os.Stat(fPath)
	if os.IsNotExist(err) {
		return false
	}
	return fileInfo.IsDir()
}

func listDirectory(dPath string) (FsEntries, error) {
	logger.Debug().Msgf("Listing directory '%s'", dPath)
	entries, err := os.ReadDir(dPath)
	if err != nil {
		return FsEntries{}, err
	}
	contents := make(FsEntries, len(entries))
	logger.Debug().Int("count", len(contents)).Msgf("Found %d item(s) for '%s'", len(contents), dPath)
	for i, e := range entries {
		relativePath := path.Join(dPath, e.Name())
		absolutePath, err := GetAbsPath(relativePath)
		if err != nil {
			return FsEntries{}, err
		}
		content := &FsEntry{
			AbsolutePath: absolutePath,
			RelativePath: relativePath,
			Name:         e.Name(),
			IsDir:        e.IsDir(),
		}
		contents[i] = content
	}
	return contents, nil
}

func listEntries(entires []*FsEntry, maxDepth int, depth int) (FsEntries, error) {
	contents := FsEntries{}
	for _, e := range entires {
		logger.Debug().Int("depth", depth).Int("maxDepth", maxDepth).Str("absPath", e.RelativePath).Msgf("Listing entries for '%s'", e.RelativePath)
		contents = append(contents, e)
		if (depth >= maxDepth && maxDepth >= 0) || !e.IsDir {
			continue
		}
		subEntries, err := listDirectory(e.RelativePath)
		if err != nil {
			return FsEntries{}, err
		}
		subContents, err := listEntries(subEntries, maxDepth, depth+1)
		if err != nil {
			return FsEntries{}, err
		}
		contents = append(contents, subContents...)
	}
	return contents, nil
}
