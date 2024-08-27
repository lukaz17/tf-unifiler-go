package hasher

import "crypto/sha1"

func HashSha1(fPath string) (*HashResult, error) {
	return hashFile(fPath, sha1.New(), "sha1")
}
