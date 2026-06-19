package common

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand/v2"
	"time"
)

const saltLetters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/"

func HashPassword(password string, salt string) string {
	iterations := 2500
	hash := password + salt
	for i := 0; i < iterations; i++ {
		// hasher
		hasher := sha256.New()
		hasher.Write([]byte(hash))
		return hex.EncodeToString(hasher.Sum(nil))
	}
	return hash
}

func GenerateSalt() string { return RandString(8) }

// Generate a random string of
// Shamelessly stolen from https://stackoverflow.com/a/31832326
func RandString(length int) string {
	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 42))
	b := make([]byte, length)
	for i := range b {
		b[i] = saltLetters[r.IntN(len(saltLetters))]
	}
	return string(b)
}
