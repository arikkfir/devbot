package util

import (
	"encoding/base64"
	"math/rand"
	"strings"
)

var (
	hashLetters = []rune("abcdefghijklmnopqrstuvwxyz")
)

func RandomHash(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = hashLetters[rand.Intn(len(hashLetters))]
	}
	return string(b)
}

func K8sCompatibleValueHash(v string) string {
	b := []byte(v)
	b64 := base64.StdEncoding.EncodeToString(b)
	return strings.Trim(b64, "-=")
}
