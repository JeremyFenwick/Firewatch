package mobinthemiddle

import (
	"strings"
	"unicode"
)

const injectionWallet = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"

// MotmAttack replaces valid Boguscoin addresses with Tony's address
func MotmAttack(content string) string {
	words := strings.Split(content, " ")

	for i, word := range words {
		if isValidAddress(word) {
			words[i] = injectionWallet
		}
	}

	return strings.Join(words, " ")
}

// Improved address validation
func isValidAddress(content string) bool {
	// Check length requirement
	if len(content) < 26 || len(content) > 35 {
		return false
	}

	// Check first character
	if content[0] != '7' {
		return false
	}

	// Check all characters are alphanumeric
	for _, c := range content {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return false
		}
	}

	return true
}
