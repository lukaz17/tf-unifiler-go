package hasher

import (
	"crypto/md5"

	"golang.org/x/crypto/md4"
)

func HashMd4(fPath string) (*HashResult, error) {
	return hashFile(fPath, md4.New(), "md4")
}

func HashMd5(fPath string) (*HashResult, error) {
	return hashFile(fPath, md5.New(), "md5")
}
