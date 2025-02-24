package model

import "time"

type RequestPayload struct {
	Limit    int       `json:"limit"`
	Start    int       `json:"start"`
	FromTime time.Time `json:"from_time"`
	ToTime   time.Time `json:"to_time"`
}
