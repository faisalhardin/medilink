package product

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	MST_PRODUCT = "mdl_mst_product"
)

const (
	WrapErrMsgPrefix              = "Conn."
	WrapMsgInsertMstProduct       = WrapErrMsgPrefix + "InsertMstProduct"
	WrapMsgFindMstProductByParams = WrapErrMsgPrefix + "FindMstProductByParams"
	WrapMsgUpdateMstProduct       = WrapErrMsgPrefix + "UpdateMstProduct"
	WrapMsgDeleteMstProduct       = WrapErrMsgPrefix + "DeleteMstProduct"
)

type Conn struct {
	DB *xormlib.DBConnect
}

func NewInstitutionDB(conn *Conn) *Conn {
	return conn
}

func (c *Conn) InsertMstProduct(ctx context.Context, institution *model.MstProduct) (err error) {
	session := c.DB.MasterDB

	_, err = session.Table(MST_PRODUCT).InsertOne(institution)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertMstProduct)
		return
	}

	return
}

func (c *Conn) FindMstProductByParams(ctx context.Context, request model.MstProduct) (products []model.MstProduct, err error) {
	session := c.DB.SlaveDB.Table(MST_PRODUCT).Alias("mmp")

	if request.ID > 0 {
		session.Where("mmp.id = ?", request.ID)
	}

	if len(request.Name) > 0 {
		session.Where(fmt.Sprintf("mmp.name ilike '%%%s%%'", request.Name))
	}

	err = session.Find(&products)
	if err != nil {
		err = errors.Wrap(err, WrapMsgFindMstProductByParams)
		return
	}

	return
}

func (c *Conn) UpdateMstProduct(ctx context.Context, mstProduct *model.MstProduct) (err error) {

	session := c.DB.MasterDB.Table(MST_PRODUCT)

	_, err = session.
		ID(mstProduct.ID).
		Cols("name", "description").
		Update(mstProduct)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateMstProduct)
		return
	}

	return
}

func (c *Conn) DeleteMstProduct(ctx context.Context, mstProduct *model.MstProduct) (err error) {

	session := c.DB.MasterDB.Table(MST_PRODUCT)

	_, err = session.
		ID(mstProduct.ID).
		Delete(mstProduct)
	if err != nil {
		err = errors.Wrap(err, WrapMsgDeleteMstProduct)
		return
	}

	return
}
