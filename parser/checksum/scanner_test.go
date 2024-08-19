package checksum

import (
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
	var tests = []struct {
		name string
		s    string
		tok  token
		lit  string
	}{
		{"EOF", "", EOF, string(eof)},
		{"Whitespace", " xyz", SPACE, " "},
		{"Long Whitespace", "   xyz", SPACE, "   "},
		{"Tab", "\txyz", SPACE, "\t"},
		{"Multi Whitespace", " \t xyz", SPACE, " \t "},
		{"Carriage Return", "\r", CR, "\r"},
		{"Carriage Return", "\r\n", CR, "\r"},
		{"Line Feed", "\n", LF, "\n"},
		{"Asterisk", "****", ASTERISK, "*"},
		{"Percent", "!@#$%^&()\\/. qwert", WORD, "!@#$%^&()\\/."},
		{"ASCII", "abcdef0123456789 qwert", WORD, "abcdef0123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.s))
			tok, lit := s.Scan()
			if tt.tok != tok {
				t.Errorf("Token mismatch. Expected %d Acutual %d", tt.tok, tok)
			} else if tt.lit != lit {
				t.Errorf("Literal mismatch. Expected '%s' Actual '%s'", tt.lit, lit)
			}
		})
	}
}
