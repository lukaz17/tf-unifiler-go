package filesystem

import (
	"reflect"
	"testing"
)

func TestList(t *testing.T) {
	prepareTests()

	tests := []struct {
		name    string
		files   []string
		results []string
	}{
		{"files only", []string{"file.go"}, []string{"file.go"}},
		{"directories only", []string{"../cmd"}, []string{"../cmd", "../cmd/args.go"}},
		{"file and directories", []string{".", "file.go"}, []string{".", "directory.go", "directory_test.go", "file.go", "file_test.go", "file.go"}},
		{"two dots", []string{"../.tests"}, []string{
			"../.tests", "../.tests/basic", "../.tests/basic/1-helloworld.md",
			"../.tests/module", "../.tests/module/fmt", "../.tests/module/fmt/1-printf.md", "../.tests/module/fmt/2-errorf.md",
			"../.tests/module/http", "../.tests/module/http/1-get.md", "../.tests/module/http/2-post.md", "../.tests/module/http/3-put.md", "../.tests/module/http/4-delete.md",
			"../.tests/module/readme.md", "../.tests/readme.md",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, err := List(tt.files, true)
			fPaths := GetPaths(contents)
			if !reflect.DeepEqual(fPaths, tt.results) {
				t.Error(err)
				t.Errorf("Wrong file listing. Expected '%s' Actual '%s'", tt.results, fPaths)
			}
		})
	}
}

func prepareTests() {
	dirs := []string{
		"../.tests",
		"../.tests/basic",
		"../.tests/module",
		"../.tests/module/fmt",
		"../.tests/module/http",
	}
	files := []string{
		"../.tests/readme.md",
		"../.tests/basic/1-helloworld.md",
		"../.tests/module/readme.md",
		"../.tests/module/fmt/1-printf.md",
		"../.tests/module/fmt/2-errorf.md",
		"../.tests/module/http/1-get.md",
		"../.tests/module/http/2-post.md",
		"../.tests/module/http/3-put.md",
		"../.tests/module/http/4-delete.md",
	}
	for _, d := range dirs {
		if !IsExist(d) {
			CreateDirectory(d)
		}
	}
	for _, f := range files {
		if !IsExist(f) {
			WriteLines(f, []string{})
		}
	}
}
