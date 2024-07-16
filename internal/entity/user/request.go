package user

type UserRequest struct {
	ID    int64  `schema:"id"`
	Email string `schema:"email" validate:"omitempty,email"`
}
