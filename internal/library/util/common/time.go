package common

import "time"

// DateOnly truncates t to UTC midnight (date only). Zero time stays zero.
func DateOnly(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// PeriodsOverlap reports whether two inclusive date ranges overlap (date-only, UTC).
func PeriodsOverlap(startA, endA, startB, endB time.Time) bool {
	aStart, aEnd := DateOnly(startA), DateOnly(endA)
	bStart, bEnd := DateOnly(startB), DateOnly(endB)
	return !aStart.After(bEnd) && !aEnd.Before(bStart)
}
