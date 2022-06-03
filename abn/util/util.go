package util

import (
	"math/rand"
	"time"
)

const (
	code = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// RandomString returns a random string of length characters
func RandomString(length int) string {
	ran_str := make([]byte, length)
	// Generating Random string
	for i := 0; i < length; i++ {
		ran_str[i] = code[rand.Intn(52)]
	}
	return string(ran_str)
}

func init() {
	rand.Seed(time.Now().Unix())
}
