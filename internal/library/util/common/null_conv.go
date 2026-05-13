package common

import "github.com/volatiletech/null/v8"

// NullableString returns a *string only when s is non-empty; nil otherwise.
func NullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// NullInt16ToPointer converts null.Int16 to *int16.
// Returns nil if the null.Int16 is not valid.
func NullInt16ToPointer(n null.Int16) *int16 {
	if !n.Valid {
		return nil
	}
	return &n.Int16
}

// NullFloat32ToPointer converts null.Float32 to *float32.
// Returns nil if the null.Float32 is not valid.
func NullFloat32ToPointer(n null.Float32) *float32 {
	if !n.Valid {
		return nil
	}
	return &n.Float32
}

// NullBoolToPointer converts null.Bool to *bool.
// Returns nil if the null.Bool is not valid.
func NullBoolToPointer(n null.Bool) *bool {
	if !n.Valid {
		return nil
	}
	return &n.Bool
}
