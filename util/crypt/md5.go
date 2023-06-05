package crypt

import (
	"crypto/md5"
	"encoding/hex"
)

func CalMd5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
