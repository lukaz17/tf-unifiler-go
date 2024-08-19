package hasher

import (
	"crypto/sha256"
	"crypto/sha512"
)

func HashSha224(fPath string) (*HashResult, error) {
	return hashFile(fPath, sha256.New224(), "sha224")
}

func HashSha256(fPath string) (*HashResult, error) {
	return hashFile(fPath, sha256.New(), "sha256")
}

func HashSha384(fPath string) (*HashResult, error) {
	return hashFile(fPath, sha512.New384(), "sha384")
}

func HashSha512(fPath string) (*HashResult, error) {
	return hashFile(fPath, sha512.New(), "sha512")
}
