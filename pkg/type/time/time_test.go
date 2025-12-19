package customtime

import (
	"reflect"
	"testing"
	"time"
)

func TestTimeConverter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     Time
		wantErr  bool
		checkNil bool // If true, expect empty reflect.Value
	}{
		{
			name:  "RFC3339 format",
			input: "2024-01-15T10:30:00Z",
			want: Time{
				Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:  "RFC3339 with timezone offset",
			input: "2024-01-15T10:30:00+07:00",
			want: Time{
				Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.FixedZone("+07:00", 7*3600)),
			},
			wantErr: false,
		},
		{
			name:  "RFC3339 with nanoseconds",
			input: "2024-01-15T10:30:00.123456789Z",
			want: Time{
				Time: time.Date(2024, 1, 15, 10, 30, 0, 123456789, time.UTC),
			},
			wantErr: false,
		},
		{
			name:  "MySQL datetime format",
			input: "2024-01-15 10:30:00",
			// Parsed as UTC then converted to Local, so we check location and date components
			want: Time{
				Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC).In(time.Local),
			},
			wantErr: false,
		},
		{
			name:  "MySQL datetime format with different time",
			input: "2023-12-25 23:59:59",
			// Parsed as UTC then converted to Local
			want: Time{
				Time: time.Date(2023, 12, 25, 23, 59, 59, 0, time.UTC).In(time.Local),
			},
			wantErr: false,
		},
		{
			name:  "Date only format",
			input: "2024-01-15",
			// Parsed as UTC then converted to Local
			want: Time{
				Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC).In(time.Local),
			},
			wantErr: false,
		},
		{
			name:  "Date only format - different date",
			input: "2023-12-25",
			// Parsed as UTC then converted to Local
			want: Time{
				Time: time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC).In(time.Local),
			},
			wantErr: false,
		},
		{
			name:     "Invalid format",
			input:    "invalid-date-format",
			want:     Time{},
			wantErr:  false,
			checkNil: true, // Should return empty reflect.Value
		},
		{
			name:     "Empty string",
			input:    "",
			want:     Time{},
			wantErr:  false,
			checkNil: true,
		},
		{
			name:     "Partial date",
			input:    "2024-01",
			want:     Time{},
			wantErr:  false,
			checkNil: true,
		},
		{
			name:     "Wrong format - slash separated",
			input:    "2024/01/15",
			want:     Time{},
			wantErr:  false,
			checkNil: true,
		},
		{
			name:  "RFC3339 - edge case: start of epoch",
			input: "1970-01-01T00:00:00Z",
			want: Time{
				Time: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:  "Date only - edge case: leap year",
			input: "2024-02-29",
			// Parsed as UTC then converted to Local
			want: Time{
				Time: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC).In(time.Local),
			},
			wantErr: false,
		},
		{
			name:     "Date only - invalid leap year",
			input:    "2023-02-29",
			want:     Time{},
			wantErr:  false,
			checkNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimeConverter(tt.input)

			if tt.checkNil {
				// For invalid formats, should return empty reflect.Value
				if !got.IsValid() || got.IsZero() {
					return // Expected behavior
				}
				t.Errorf("TimeConverter() = %v, expected empty reflect.Value for invalid input", got)
				return
			}

			// Check if reflect.Value is valid
			if !got.IsValid() {
				t.Errorf("TimeConverter() returned invalid reflect.Value for input: %s", tt.input)
				return
			}

			// Extract the Time value from reflect.Value
			if !got.CanInterface() {
				t.Errorf("TimeConverter() returned reflect.Value that cannot interface")
				return
			}

			gotTime, ok := got.Interface().(Time)
			if !ok {
				t.Errorf("TimeConverter() returned type %T, expected Time", got.Interface())
				return
			}

			// Compare times (allowing for small differences due to timezone handling)
			if !timesEqual(gotTime.Time, tt.want.Time) {
				t.Errorf("TimeConverter() = %v, want %v", gotTime.Time, tt.want.Time)
			}
		})
	}
}

// timesEqual compares two time.Time values, accounting for potential timezone differences
func timesEqual(t1, t2 time.Time) bool {
	// Compare Unix timestamps to avoid timezone issues
	return t1.Unix() == t2.Unix() && t1.Nanosecond() == t2.Nanosecond()
}

func TestTimeConverter_MySQLDatetimeFormat_LocationHandling(t *testing.T) {
	// Test that MySQL datetime format converts UTC to Local
	input := "2024-01-15 10:30:00"

	result := TimeConverter(input)
	if !result.IsValid() {
		t.Fatal("TimeConverter() returned invalid reflect.Value")
	}

	gotTime, ok := result.Interface().(Time)
	if !ok {
		t.Fatalf("TimeConverter() returned type %T, expected Time", result.Interface())
	}

	// The time should be in Local timezone (not UTC)
	if gotTime.Time.Location() != time.Local {
		t.Errorf("TimeConverter() MySQL format should convert to Local timezone, got %v", gotTime.Time.Location())
	}

	// Verify the date components are correct (time will be adjusted for timezone)
	if gotTime.Time.Year() != 2024 || gotTime.Time.Month() != 1 || gotTime.Time.Day() != 15 {
		t.Errorf("TimeConverter() date components incorrect: got %v", gotTime.Time)
	}

	// The time was parsed as UTC (10:30 UTC) then converted to Local
	// So we verify it's in Local timezone and the UTC time matches the input
	utcTime := gotTime.Time.UTC()
	if utcTime.Hour() != 10 || utcTime.Minute() != 30 || utcTime.Second() != 0 {
		t.Errorf("TimeConverter() UTC time components should match input: got %v (UTC: %v)", gotTime.Time, utcTime)
	}
}

func TestTimeConverter_DateOnlyFormat_LocationHandling(t *testing.T) {
	// Test that date-only format converts UTC to Local
	input := "2024-01-15"

	result := TimeConverter(input)
	if !result.IsValid() {
		t.Fatal("TimeConverter() returned invalid reflect.Value")
	}

	gotTime, ok := result.Interface().(Time)
	if !ok {
		t.Fatalf("TimeConverter() returned type %T, expected Time", result.Interface())
	}

	// The time should be in Local timezone (not UTC)
	if gotTime.Time.Location() != time.Local {
		t.Errorf("TimeConverter() date-only format should convert to Local timezone, got %v", gotTime.Time.Location())
	}

	// Verify the date components are correct
	if gotTime.Time.Year() != 2024 || gotTime.Time.Month() != 1 || gotTime.Time.Day() != 15 {
		t.Errorf("TimeConverter() date components incorrect: got %v", gotTime.Time)
	}

	// The time was parsed as UTC (00:00:00 UTC) then converted to Local
	// So we verify the UTC time is 00:00:00
	utcTime := gotTime.Time.UTC()
	if utcTime.Hour() != 0 || utcTime.Minute() != 0 || utcTime.Second() != 0 {
		t.Errorf("TimeConverter() UTC time should be 00:00:00 for date-only format, got %v (UTC: %v)", gotTime.Time, utcTime)
	}
}

func TestTimeConverter_RFC3339_TimezonePreservation(t *testing.T) {
	// Test that RFC3339 format preserves the original timezone
	tests := []struct {
		name           string
		input          string
		expectedUTC    bool
		expectedOffset string // Expected offset string like "+07:00" or "-05:00"
	}{
		{
			name:        "UTC timezone",
			input:       "2024-01-15T10:30:00Z",
			expectedUTC: true,
		},
		{
			name:           "Positive offset",
			input:          "2024-01-15T10:30:00+07:00",
			expectedUTC:    false,
			expectedOffset: "+07:00",
		},
		{
			name:           "Negative offset",
			input:          "2024-01-15T10:30:00-05:00",
			expectedUTC:    false,
			expectedOffset: "-05:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeConverter(tt.input)
			if !result.IsValid() {
				t.Fatal("TimeConverter() returned invalid reflect.Value")
			}

			gotTime, ok := result.Interface().(Time)
			if !ok {
				t.Fatalf("TimeConverter() returned type %T, expected Time", result.Interface())
			}

			// Check that timezone is preserved for RFC3339
			gotLocation := gotTime.Time.Location()

			// For UTC, compare directly
			if tt.expectedUTC {
				if gotLocation != time.UTC {
					t.Errorf("TimeConverter() RFC3339 should preserve UTC timezone, got %v", gotLocation)
				}
			} else {
				// For other timezones, compare the offset
				gotOffset := gotTime.Time.Format("-07:00")
				if gotOffset != tt.expectedOffset {
					t.Errorf("TimeConverter() RFC3339 should preserve timezone offset, got %v, want %v", gotOffset, tt.expectedOffset)
				}
			}
		})
	}
}

func TestTimeConverter_ReflectValueType(t *testing.T) {
	// Test that the returned reflect.Value has the correct type
	input := "2024-01-15T10:30:00Z"
	result := TimeConverter(input)

	if !result.IsValid() {
		t.Fatal("TimeConverter() returned invalid reflect.Value")
	}

	// Check the type
	expectedType := reflect.TypeOf(Time{})
	if result.Type() != expectedType {
		t.Errorf("TimeConverter() returned type %v, expected %v", result.Type(), expectedType)
	}

	// Verify it's not a pointer
	if result.Kind() == reflect.Ptr {
		t.Error("TimeConverter() should return Time value, not pointer")
	}
}
