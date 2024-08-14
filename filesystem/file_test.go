package filesystem

import (
	"reflect"
	"testing"
)

func TestList(t *testing.T) {
	tests := []struct {
		name    string
		files   []string
		results []string
	}{
		{"TestList: files only", []string{"file.go"}, []string{"file.go"}},
		{"TestList: directories only", []string{"../cmd"}, []string{"../cmd", "../cmd/args.go"}},
		{"TestList: file and directories", []string{".", "file.go"}, []string{".", "directory.go", "file.go", "file_test.go", "file.go"}},
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
