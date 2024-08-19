package hasher

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/tforceaio/tf-unifiler-go/extension"
)

type HashResult struct {
	Path      string
	Algorithm string
	Hash      []byte
}

func Hash(fPath string, algorithms []string) ([]*HashResult, error) {
	fHandle, err := os.Open(fPath)
	if err != nil {
		return []*HashResult{}, err
	}
	defer fHandle.Close()

	results := make([]*HashResult, len(algorithms))
	hashers := make([]hash.Hash, len(algorithms))
	for i, a := range algorithms {
		switch a {
		case "md5":
			results[i] = &HashResult{
				Path:      fPath,
				Algorithm: "md5",
			}
			hashers[i] = md5.New()
		case "sha1":
			results[i] = &HashResult{
				Path:      fPath,
				Algorithm: "sha1",
			}
			hashers[i] = sha1.New()
		case "sha256":
			results[i] = &HashResult{
				Path:      fPath,
				Algorithm: "sha256",
			}
			hashers[i] = sha256.New()
		default:
			return []*HashResult{}, fmt.Errorf("unsupported hash algorithm: '%s'", a)
		}
	}

	bufSize := getBufferSize(fHandle)
	buf := make([]byte, bufSize)
	written := int64(0)
	for {
		nread, eread := fHandle.Read(buf)
		if nread > 0 {
			nwrite := 0
			var ewrite error = nil
			for _, hasher := range hashers {
				nwrite, ewrite = hasher.Write(buf[0:nread])
			}
			if nwrite < 0 || nread < nwrite {
				nwrite = 0
				if ewrite == nil {
					ewrite = errors.New("cannot write to hasher")
				}
			}
			written += int64(nwrite)
			if ewrite != nil {
				err = ewrite
				break
			}
			if nread != nwrite {
				err = fmt.Errorf("read and write data mismatch %d %d", nread, nwrite)
				break
			}
		}
		if eread != nil {
			if eread != io.EOF {
				err = eread
			}
			break
		}
	}
	if err != nil {
		return []*HashResult{}, err
	}

	for i, h := range hashers {
		results[i].Hash = h.Sum(nil)
	}
	logger.Info().Array("algos", extension.StringSlice(algorithms)).Str("file", fPath).Int("size", int(written)).Msgf("Hashed '%s' (%d bytes)", fPath, written)
	return results, nil
}

func getBufferSize(src io.Reader) int {
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	return size
}

func hashFile(fPath string, hasher hash.Hash, algo string) (*HashResult, error) {
	fHandle, err := os.Open(fPath)
	if err != nil {
		return nil, err
	}
	defer fHandle.Close()

	bufSize := getBufferSize(fHandle)
	buf := make([]byte, bufSize)
	written := int64(0)
	for {
		nread, eread := fHandle.Read(buf)
		if nread > 0 {
			nwrite, ewrite := hasher.Write(buf[0:nread])
			if nwrite < 0 || nread < nwrite {
				nwrite = 0
				if ewrite == nil {
					ewrite = errors.New("cannot write to hasher")
				}
			}
			written += int64(nwrite)
			if ewrite != nil {
				err = ewrite
				break
			}
			if nread != nwrite {
				err = fmt.Errorf("read and write data mismatch %d %d", nread, nwrite)
				break
			}
		}
		if eread != nil {
			if eread != io.EOF {
				err = eread
			}
			break
		}
	}
	if err != nil {
		return nil, err
	}

	result := &HashResult{
		Path:      fPath,
		Algorithm: algo,
		Hash:      hasher.Sum(nil),
	}
	logger.Info().Str("algo", algo).Str("file", fPath).Hex("hash", result.Hash).Int("size", int(written)).Msgf("Hashed '%s' (%d bytes)", fPath, written)
	return result, nil
}
