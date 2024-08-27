package myrand

import (
	"math/rand/v2"
	"strings"
)

const (
	AlphabetLowercase    = "abcdefghijklmnopqrstuvwxyz"
	AlphabetUppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AlphabetNumbers      = "0123456789"
	AlphabetAlphaNumeric = AlphabetLowercase + AlphabetUppercase + AlphabetNumbers
)

func AlphaNumeric(length int) string {
	return buildString(AlphabetAlphaNumeric, length)
}

func buildString(alphabet string, length int) string {
	var sb strings.Builder

	for range length {
		sb.WriteByte(alphabet[rand.Int()%len(alphabet)])
	}

	return sb.String()
}
