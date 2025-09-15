package shortid

import (
	"crypto/rand"
	"math/big"
)

const (
	// Base58 alphabet (Bitcoin style - excludes 0, O, I, l)
	base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	base58Length   = len(base58Alphabet)
)

// GenerateShortID generates a random Base58 encoded short ID of specified length
func GenerateShortID(length int) (string, error) {
	if length <= 0 {
		length = 8
	}

	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(base58Length)))
		if err != nil {
			return "", err
		}
		result[i] = base58Alphabet[n.Int64()]
	}

	return string(result), nil
}

// GenerateShortIDWithPrefix generates a short ID with a prefix
func GenerateShortIDWithPrefix(prefix string, length int) (string, error) {
	shortID, err := GenerateShortID(length)
	if err != nil {
		return "", err
	}
	return prefix + shortID, nil
}

// ValidateShortID checks if a string is a valid Base58 short ID
func ValidateShortID(id string) bool {
	if len(id) == 0 {
		return false
	}

	for _, char := range id {
		if !contains(base58Alphabet, char) {
			return false
		}
	}
	return true
}

// contains checks if a character exists in the alphabet
func contains(alphabet string, char rune) bool {
	for _, c := range alphabet {
		if c == char {
			return true
		}
	}
	return false
}
