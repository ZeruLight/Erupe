package token

import (
	"math/rand"
	"time"
)

var RNG = NewRNG()

// Generate returns an alphanumeric token of specified length
func Generate(length int) string {
	var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = chars[RNG.Intn(len(chars))]
	}
	return string(b)
}

// NewRNG returns a new NewRNG generator
func NewRNG() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
