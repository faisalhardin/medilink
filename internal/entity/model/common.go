package model

type RequestPayload struct {
	Limit int `json:"limit"`
	Start int `json:"start"`
}
