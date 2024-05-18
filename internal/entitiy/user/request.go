package user

type UserRequest struct {
	ID    int64  `schema:"id"`
	Email string `json:"email" validate:"omitempty,email"`
}
