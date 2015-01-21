package file_control

import (
	"encoding/hex"
	"crypto/md5"
)

const CHUNK_SIZE = 4

func Hash(b []byte) string {
	byteArray := md5.Sum(b)
	return hex.EncodeToString(byteArray[:])
}
