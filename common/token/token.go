package token

import (
	"math/rand"
	"time"
)

// Generate returns an alphanumeric token of specified length
func Generate(length int) string {
	rng := RNG()
	var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	return string(b)
}

// RNG returns a new RNG generator
func RNG() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
