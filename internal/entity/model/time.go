package model

import (
	"encoding/json"
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

func (t *Time) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	date, err := time.Parse(time.RFC3339, s)
	if err == nil {
		*t = Time(date)
		return nil
	}

	date, err = time.Parse("2006-01-02", s)
	if err == nil {
		*t = Time(date)
		return nil
	}
	return err
}
