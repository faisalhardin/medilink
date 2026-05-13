package xorm

import "time"

// This file holds small type-coercion helpers for values returned by
// xorm's (*Session).QueryInterface(), which yields map[string]interface{}
// where Go types are whatever the underlying pq driver chose (often
// []byte, int64, or time.Time). Struct-scanning paths (Find / Get) do
// not need these — use them only when reading back columns from raw
// RETURNING queries.

// ToTime extracts a time.Time from an interface. Returns the zero value
// when the column was NULL or the driver returned an unexpected type.
func ToTime(v interface{}) time.Time {
	if t, ok := v.(time.Time); ok {
		return t
	}
	return time.Time{}
}

// ToInt16Ptr coerces the dynamic numeric type returned by pq/xorm
// (usually int64 for SMALLINT) into a *int16. Returns nil for NULL or
// an unrecognised type so callers can round-trip nullable columns.
func ToInt16Ptr(v interface{}) *int16 {
	switch n := v.(type) {
	case int64:
		x := int16(n)
		return &x
	case int16:
		x := n
		return &x
	case int:
		x := int16(n)
		return &x
	case nil:
		return nil
	default:
		return nil
	}
}
