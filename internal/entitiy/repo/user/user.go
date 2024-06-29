package user

import (
	"context"

	"github.com/faisalhardin/auth-vessel/internal/entitiy/user"
)

type UserDB interface {
	InserUser(ctx context.Context, user *user.User) (err error)
	GetUserByParams(ctx context.Context, params user.User) (resp user.User, found bool, err error)
}
