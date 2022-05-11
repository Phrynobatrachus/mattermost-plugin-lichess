package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func generateSecret(enc base64.Encoding, len int) (string, error) {
	b := make([]byte, 256)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	s := enc.EncodeToString(b)

	s = s[:len]

	return s, nil
}

func genCodeChallengeS256(s string) string {
	s256 := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(s256[:])
}
