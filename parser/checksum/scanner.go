package checksum

import (
	"bufio"
	"bytes"
	"io"
)

var eof = rune(0)

type scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r)}
}

func (s *scanner) Scan() (tok token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isWord(ch) {
		s.unread()
		return s.scanWord()
	}

	switch ch {
	case '*':
		return ASTERISK, "*"
	case '\r':
		return CR, "\r"
	case '\n':
		return LF, "\n"
	case eof:
		return EOF, string(eof)
	}

	return INVALID, string(ch)
}

func (s *scanner) scanWhitespace() (tok token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); isWhitespace(ch) {
			buf.WriteRune(ch)
		} else {
			s.unread()
			break
		}
	}

	return SPACE, buf.String()
}

func (s *scanner) scanWord() (tok token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); isWord(ch) {
			_, _ = buf.WriteRune(ch)
		} else {
			s.unread()
			break
		}
	}

	return WORD, buf.String()
}

func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *scanner) unread() {
	_ = s.r.UnreadRune()
}

func isAsterisk(ch rune) bool {
	return ch == '*'
}

func isEndline(ch rune) bool {
	return ch == '\r' || ch == '\n'
}

func isEndOfFile(ch rune) bool {
	return ch == eof
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func isWord(ch rune) bool {
	return !isAsterisk(ch) && !isEndline(ch) && !isEndOfFile(ch) && !isWhitespace(ch)
}
