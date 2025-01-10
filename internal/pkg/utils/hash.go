package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// func HashKey(k string) string {
// 	h := sha256.New()
// 	_, _ = h.Write([]byte(k))
// 	return hex.EncodeToString(h.Sum(nil))
// }

func HashKeyWithHMAC(key, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

type Hasher struct {
	secret string
}

func NewHasher(secret string) (*Hasher, error) {
	if len(secret) < 30 {
		return nil, fmt.Errorf("secret is too short")
	}
	return &Hasher{secret: secret}, nil
}

func (h *Hasher) HashKey(key string) string {
	return HashKeyWithHMAC(key, h.secret)
}
