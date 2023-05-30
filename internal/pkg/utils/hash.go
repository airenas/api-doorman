package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashKey(k string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(k))
	return hex.EncodeToString(h.Sum(nil))
}
