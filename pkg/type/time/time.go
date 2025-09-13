package customtime

import (
	"reflect"
	"time"
)

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02T15:04:05.000-0700"`, string(b))
	if err == nil {
		t.Time = date
		return
	}

	date, err = time.Parse(`"2006-01-02"`, string(b))
	if err == nil {
		// Set default location to avoid "missing Location" error
		if date.Location() == time.UTC {
			date = date.In(time.Local)
		}
		t.Time = date
		return
	}
	return err
}

var TimeConverter = func(value string) reflect.Value {
	if v, err := time.Parse(time.RFC3339, value); err == nil {
		t := Time{
			Time: v,
		}
		return reflect.ValueOf(t)
	}

	if v, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		// Set default location to avoid "missing Location" error
		if v.Location() == time.UTC {
			v = v.In(time.Local)
		}
		t := Time{
			Time: v,
		}
		return reflect.ValueOf(t)
	}

	if v, err := time.Parse("2006-01-02", value); err == nil {
		// Set default location to avoid "missing Location" error
		if v.Location() == time.UTC {
			v = v.In(time.Local)
		}
		t := Time{
			Time: v,
		}
		return reflect.ValueOf(t)
	}
	return reflect.Value{}
}
