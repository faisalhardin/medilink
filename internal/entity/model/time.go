package model

import (
	"reflect"
	"time"
)

// Time special binding for time, avoid tampering default binding to time.Time
type Time time.Time

// Time convert to time.Time struct
func (t Time) Time() time.Time {
	return time.Time(t)
}

func TimeConverter(value string) reflect.Value {
	if v, err := time.Parse(time.RFC3339, value); err == nil {
		return reflect.ValueOf(v)
	}

	if v, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return reflect.ValueOf(v)
	}

	if v, err := time.Parse("2006-01-02", value); err == nil {
		return reflect.ValueOf(v)
	}

	return reflect.Value{} // this is the same as the private const invalidType
}