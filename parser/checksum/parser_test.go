package checksum

import (
	"strings"
	"testing"

	"github.com/tforceaio/tf-unifiler-go/extension"
)

func TestParserItemCount(t *testing.T) {
	var tests = []struct {
		name    string
		content string
		count   int
	}{
		{"0 item", "", 0},
		{"1 item", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod", 1},
		{"2 items", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod\nca47868bca0d531a275f20e99eb04ba1 go.sum", 2},
		{"CRLF", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod\r\nca47868bca0d531a275f20e99eb04ba1 go.sum", 2},
		{"Trailing CRLF", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod\r\nca47868bca0d531a275f20e99eb04ba1 go.sum\n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(strings.NewReader(tt.content))
			items, _ := p.Parse()
			if len(items) != tt.count {
				t.Errorf("Invalid number of items. Expected %d. Actual %d.", tt.count, len(items))
			}
		})
	}
}

func TestParserSyntax(t *testing.T) {
	var tests = []struct {
		name       string
		content    string
		path       string
		binaryMode bool
		hash       string
	}{
		{"Text mode", "a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod", "go.mod", false, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Binary mode", "a3c51dd48bf7fabbbd354bd4e16b0ec1 *go.mod", "go.mod", true, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Multiple space", "a3c51dd48bf7fabbbd354bd4e16b0ec1    go.mod", "go.mod", false, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Path with multiple segment", "a3c51dd48bf7fabbbd354bd4e16b0ec1 filesystem/file.go", "filesystem/file.go", false, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Path with space #1", "a3c51dd48bf7fabbbd354bd4e16b0ec1 file system/file.go", "file system/file.go", false, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
		{"Path with space #2", "a3c51dd48bf7fabbbd354bd4e16b0ec1 *file system/file.go", "file system/file.go", true, "a3c51dd48bf7fabbbd354bd4e16b0ec1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(strings.NewReader(tt.content))
			items, _ := p.Parse()
			item := items[0]
			if item.Path != tt.path {
				t.Errorf("Wrong path. Expected '%s'. Actual '%s'.", tt.path, item.Path)
			}
			if item.BinaryMode != tt.binaryMode {
				t.Errorf("Wrong mode. Expected %t. Actual %t.", tt.binaryMode, item.BinaryMode)
			}
			if item.Hash != tt.hash {
				t.Errorf("Wrong hash. Expected '%s'. Actual '%s'.", tt.hash, item.Hash)
			}
		})
	}
}

func TestParserError(t *testing.T) {
	var tests = []struct {
		name    string
		content string
		err     string
	}{
		{"hash only", "a3c51dd48bf7fabbbd354bd4e16b0ec1", "invalid token. expected whitespace actual '\x00'"},
		{"hash only", "a3c51dd48bf7fabbbd354bd4e16b0ec1\n", "invalid token. expected whitespace actual '\n'"},
		{"missing path", "a3c51dd48bf7fabbbd354bd4e16b0ec1  ", "invalid token. expected whitespace actual '\x00'"},
		{"missing path", "a3c51dd48bf7fabbbd354bd4e16b0ec1  \n", "invalid token. expected whitespace actual '\n'"},
		{"missing path", "a3c51dd48bf7fabbbd354bd4e16b0ec1  *", "invalid token. expected path actual '\x00'"},
		{"missing path", "a3c51dd48bf7fabbbd354bd4e16b0ec1  *\n", "invalid token. expected path actual '\n'"},
		{"path only", "go.mod", "invalid token. expected whitespace actual '\x00'"},
		{"path only", "go.mod\n", "invalid token. expected whitespace actual '\n'"},
		{"missing hash", " go.mod", "invalid token. expected hash actual ' '"},
		{"path only", "*go.mod", "invalid token. expected hash actual '*'"},
		{"path only", "*go.mod\n", "invalid token. expected hash actual '*'"},
		{"missing hash", " *go.mod\n", "invalid token. expected hash actual ' '"},
		{"space before path", "a3c51dd48bf7fabbbd354bd4e16b0ec1  * go.mod", "invalid token. expected path actual ' '"},
		{"space before path", "a3c51dd48bf7fabbbd354bd4e16b0ec1  * go.mod\n", "invalid token. expected path actual ' '"},
		{"space after path", "a3c51dd48bf7fabbbd354bd4e16b0ec1  *go.mod ", "invalid token. expected path actual ' '"},
		{"space after path", "a3c51dd48bf7fabbbd354bd4e16b0ec1  *go.mod \n", "invalid token. expected path actual ' '"},
		{"space before hash", " a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod", "invalid token. expected hash actual ' '"},
		{"space before hash", " a3c51dd48bf7fabbbd354bd4e16b0ec1 go.mod\n", "invalid token. expected hash actual ' '"},
		{"space in hash", "a3c51dd48bf7fabbbd 354bd4e16b0ec1 go.mod", ""},
		{"space in hash", "a3c51dd48bf7fabbbd 354bd4e16b0ec1 go.mod\n", ""},
		{"space in hash", "a3c51dd48bf7fabbbd 354bd4e16b0ec1 *go.mod", "invalid token. expected path actual '*'"},
		{"space in hash", "a3c51dd48bf7fabbbd 354bd4e16b0ec1 *go.mod\n", "invalid token. expected path actual '*'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(strings.NewReader(tt.content))
			_, err := p.Parse()
			errs := extension.ErrString(err)
			if errs != tt.err {
				t.Errorf("wrong error. Expected %q. Actual %q.", tt.err, errs)
			}
		})
	}
}
