package randkey

import (
	"math/rand"
)

var syms = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randKey(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = syms[rand.Intn(len(syms))]
	}
	return string(b)
}

// Generate genetares new random key
func Generate() string {
	return randKey(10)
}
