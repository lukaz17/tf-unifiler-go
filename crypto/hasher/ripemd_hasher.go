package hasher

import "golang.org/x/crypto/ripemd160"

func HashRipemd160(fPath string) (*HashResult, error) {
	return hashFile(fPath, ripemd160.New(), "ripemd160")
}
