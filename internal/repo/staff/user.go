package staff

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/constant/database"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/entity/user"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	TableMstUser     = "av_mst_user"
	TableMstRole     = "av_mst_role"
	TableMapUserRole = "av_map_user_role"
)

type Conn struct {
	DB *xormlib.DBConnect
}

func New(conn Conn) Conn {
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

// function GetUserDetailByEmail run the following query:
// SELECT mms.*, jsonb_agg(json_build_object('role_id',mmr.role_id, 'name', mmr."name")) as roles
// FROM "mdl_mst_staff" AS "mms"
// INNER JOIN mdl_map_role_staff mmrs ON mmrs.id_mst_staff = mms.id
// INNER JOIN mdl_mst_role mmr ON mmr.id = mmrs.id_mst_role and mmr.delete_time is null
// WHERE ("mms"."delete_time" IS NULL)
// group by mms.id;
func (c *Conn) GetUserDetailByEmail(ctx context.Context, email string) (staff model.UserDetail, err error) {
	session := c.DB.SlaveDB.Table("mdl_mst_staff").Alias("mms")

	found, err := session.
		Join(database.SQLInner, "mdl_map_role_staff mmrs", "mmrs.id_mst_staff = mms.id").
		Join(database.SQLInner, "mdl_mst_role mmr", "mmr.id = mmrs.id_mst_role and mmr.delete_time is null").
		Select("mms.*, jsonb_agg(json_build_object('role_id',mmr.role_id, 'name', mmr.name)) as roles").
		GroupBy("mms.id").
		Get(&staff)
	if err != nil {
		err = errors.Wrap(err, "GetUserDetailByEmail")
		return
	}
	if !found {
		err = commonerr.SetNewBadRequest("User not found", "Please login with a registered email")
		return
	}

	return
}
