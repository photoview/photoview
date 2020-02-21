package utils

import "crypto/rand"

import "math/big"

import "log"

func GenerateToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	charLen := big.NewInt(int64(len(charset)))

	b := make([]byte, length)
	for i := range b {

		n, err := rand.Int(rand.Reader, charLen)
		if err != nil {
			log.Fatalf("Could not generate random number: %s\n", err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
