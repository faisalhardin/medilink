package institution

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	MST_INSTITUTION = "mdl_mst_institution"
)

const (
	WrapErrMsgPrefix            = "Conn."
	WrapMsgInsertNewInstitution = WrapErrMsgPrefix + "InsertNewInstitution"
	WrapMsgGetInstitution       = WrapErrMsgPrefix + "GetInstitution"
)

type Conn struct {
	DB *xormlib.DBConnect
}

func NewInstitutionDB(conn *Conn) *Conn {
	return conn
}

func (c *Conn) InsertNewInstitution(ctx context.Context, institution *model.Institution) (err error) {
	session := c.DB.MasterDB

	_, err = session.Table(MST_INSTITUTION).InsertOne(institution)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewInstitution)
		return
	}

	return
}

func (c *Conn) FindInstitutionByParams(ctx context.Context, request model.FindInstitutionParams) (institutions []model.Institution, err error) {
	session := c.DB.SlaveDB.Table(MST_INSTITUTION).Alias("mmi")

	if request.ID > 0 {
		session.Where("mmi.id = ?", request.ID)
	}

	if len(request.Name) > 0 {
		session.Where(fmt.Sprintf("mmi.name ilike '%%%s%%'", request.Name))
	}

	err = session.Find(&institutions)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetInstitution)
		return
	}

	return
}
