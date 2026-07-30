package hash

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// SHA256JSON marshals v to JSON and returns its SHA-256 hex digest.
func SHA256JSON(v interface{}) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum)
}
