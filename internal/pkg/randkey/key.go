package randkey

import (
	"crypto/rand"
	"math/big"
)

// Generate genetares new random key
func Generate(n int) (string, error) {
	var syms = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	l := big.NewInt(int64(len(syms)))

	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, l)
		if err != nil {
			return "", err
		}
		b[i] = syms[num.Int64()]
	}
	return string(b), nil
}
