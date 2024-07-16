package user

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/user"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
)

const (
	TableMstUser     = "av_mst_user"
	TableMstRole     = "av_mst_role"
	TableMapUserRole = "av_map_user_role"
)

type Conn struct {
	DB *xormlib.DBConnect
}

func New(conn *Conn) *Conn {
	return conn
}

func (c *Conn) InserUser(ctx context.Context, user *user.User) (err error) {
	session := c.DB.MasterDB

	_, err = session.Table("av_mst_user").Get(user)
	if err != nil {
		return
	}

	return
}

func (c *Conn) GetUserByParams(ctx context.Context, params user.User) (resp user.User, found bool, err error) {
	session := c.DB.SlaveDB.Table("av_mst_user").Alias("amu")

	if len(params.Email) > 0 {
		session.Where("amu.email = ?", params.Email)
	}

	found, err = session.Get(&resp)
	if err != nil {
		return
	}

	return
}
