package randkey

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var syms = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// Generate genetares new random key
func Generate(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = syms[rand.Intn(len(syms))]
	}
	return string(b)
}
