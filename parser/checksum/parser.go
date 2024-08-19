package checksum

import (
	"fmt"
	"io"
	"strings"
)

type Parser struct {
	s   *scanner
	buf struct {
		tok token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

func (p *Parser) Parse() ([]*ChecksumItem, error) {
	items := []*ChecksumItem{}

	for {
		item := &ChecksumItem{}
		if tok, lit := p.scan(); tok == EOF {
			break
		} else if tok == WORD {
			item.Hash = lit
		} else {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "hash", lit)
		}

		if tok, lit := p.scan(); tok != SPACE {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "whitespace", lit)
		}

		pathSlice := []string{}
		if tok, lit := p.scan(); tok == ASTERISK {
			item.BinaryMode = true
		} else if tok == WORD || tok == SPACE {
			p.unscan()
		} else {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "whitespace", lit)
		}

		if tok, lit := p.scan(); tok == SPACE {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "path", lit)
		} else {
			p.unscan()
		}

		var lastTok token
		for {
			if tok, lit := p.scan(); tok == SPACE || tok == WORD {
				lastTok = tok
				pathSlice = append(pathSlice, lit)
			} else if tok == CR || tok == LF || tok == EOF {
				p.unscan()
				item.Path = strings.Join(pathSlice, "")
				break
			} else {
				return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "path", lit)
			}
		}
		if len(pathSlice) == 0 {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "path", p.buf.lit)
		}
		if lastTok == SPACE {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expected %s actual '%s'", "path", " ")
		}

		if tok, lit := p.scan(); tok == CR {
			if tok, lit = p.scan(); tok == LF {
				items = append(items, item)
			} else {
				return []*ChecksumItem{}, fmt.Errorf("invalid token. expcted %s actual '%s'", "endline", lit)
			}
		} else if tok == LF {
			items = append(items, item)
		} else if tok == EOF {
			items = append(items, item)
			break
		} else {
			return []*ChecksumItem{}, fmt.Errorf("invalid token. expcted %s actual '%s'", "endline", lit)
		}
	}

	return items, nil
}

func (p *Parser) scan() (token, string) {
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	tok, lit := p.s.Scan()
	p.buf.tok, p.buf.lit = tok, lit
	return tok, lit
}

func (p *Parser) unscan() {
	p.buf.n = 1
}
