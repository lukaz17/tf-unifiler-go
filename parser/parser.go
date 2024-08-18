package parser

import (
	"fmt"
	"io"
	"regexp"

	"github.com/tforceaio/tf-unifiler-go/parser/checksum"
)

func ParseSha256(r io.Reader) ([]*checksum.ChecksumItem, error) {
	parser := checksum.NewParser(r)
	items, err := parser.Parse()
	if err != nil {
		return items, err
	}
	hashCheckRegex := regexp.MustCompile("^[0-9A-Fa-f]{64}$")
	for _, l := range items {
		if !hashCheckRegex.MatchString(l.Hash) {
			return []*checksum.ChecksumItem{}, fmt.Errorf("invalid SHA-256 hash '%s'", l.Hash)
		}
	}
	return items, err
}
