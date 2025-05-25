package model

import "time"

type CommonRequestPayload struct {
	Limit    int       `json:"limit" schema:"limit"`
	Offset   int       `json:"offset" schema:"offset"`
	Page     int       `json:"page" schema:"page"`
	FromTime time.Time `json:"from_time" schema:"from_time"`
	ToTime   time.Time `json:"to_time" schema:"to_time"`
}

type RangeFilterPayload struct {
	Operator string `json:"operator"`
	Value    string `json:"value"`
}
