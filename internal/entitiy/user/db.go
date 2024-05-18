package user

type User struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name" validate:"required"`
	Email string  `json:"email" validate:"email"`
	Money float64 `json:"money" validate:"max=10"`
}
