package checksum

type ChecksumItem struct {
	Hash       string
	BinaryMode bool
	Path       string
}

type token int

const (
	INVALID token = iota
	SPACE
	CR
	LF
	EOF

	ASTERISK

	WORD
)
