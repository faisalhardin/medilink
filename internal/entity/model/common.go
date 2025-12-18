package model

import customtime "github.com/faisalhardin/medilink/pkg/type/time"

type CommonRequestPayload struct {
	Limit    int             `json:"limit" schema:"limit" validate:"omitempty,min=0"`
	Offset   int             `json:"offset" schema:"offset" validate:"omitempty,min=0"`
	Page     int             `json:"page" schema:"page" validate:"omitempty,min=0"`
	FromTime customtime.Time `json:"from_time" schema:"from_time" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	ToTime   customtime.Time `json:"to_time" schema:"to_time" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	OrderBy  string          `json:"order_by" schema:"order_by"`
}

type RangeFilterPayload struct {
	Operator string `json:"operator"`
	Value    string `json:"value"`
}
