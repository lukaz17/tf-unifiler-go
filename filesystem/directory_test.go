package filesystem

import (
	"reflect"
	"testing"
)

func TestListEntires(t *testing.T) {
	prepareTests()

	tests := []struct {
		name     string
		files    []string
		maxDepth int
		results  []string
	}{
		{"multiple entries", []string{"../.tests/module/http", "../.tests/module/fmt"}, -1, []string{
			"../.tests/module/http", "../.tests/module/http/1-get.md", "../.tests/module/http/2-post.md", "../.tests/module/http/3-put.md", "../.tests/module/http/4-delete.md",
			"../.tests/module/fmt", "../.tests/module/fmt/1-printf.md", "../.tests/module/fmt/2-errorf.md",
		}},
		{"depth level 1", []string{"../.tests/module", "../.tests/basic"}, 1, []string{
			"../.tests/module", "../.tests/module/fmt", "../.tests/module/http", "../.tests/module/readme.md",
			"../.tests/basic", "../.tests/basic/1-helloworld.md",
		}},
		{"depth level 2", []string{"../.tests/module", "../.tests/basic"}, 2, []string{
			"../.tests/module", "../.tests/module/fmt", "../.tests/module/fmt/1-printf.md", "../.tests/module/fmt/2-errorf.md",
			"../.tests/module/http", "../.tests/module/http/1-get.md", "../.tests/module/http/2-post.md", "../.tests/module/http/3-put.md", "../.tests/module/http/4-delete.md",
			"../.tests/module/readme.md",
			"../.tests/basic", "../.tests/basic/1-helloworld.md",
		}},
		{"recursive", []string{"../.tests"}, -1, []string{
			"../.tests", "../.tests/basic", "../.tests/basic/1-helloworld.md",
			"../.tests/module", "../.tests/module/fmt", "../.tests/module/fmt/1-printf.md", "../.tests/module/fmt/2-errorf.md",
			"../.tests/module/http", "../.tests/module/http/1-get.md", "../.tests/module/http/2-post.md", "../.tests/module/http/3-put.md", "../.tests/module/http/4-delete.md",
			"../.tests/module/readme.md",
			"../.tests/readme.md",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := make([]*FsEntry, len(tt.files))
			for i, f := range tt.files {
				entries[i], _ = CreateEntry(f)
			}
			contents, err := listEntries(entries, tt.maxDepth, 0)
			fPaths := GetPaths(contents)
			if !reflect.DeepEqual(fPaths, tt.results) {
				t.Error(err)
				t.Errorf("Wrong file listing. Expected '%s' Actual '%s'", tt.results, fPaths)
			}
		})
	}
}
