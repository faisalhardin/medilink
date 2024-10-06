package institution

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	TrxInstitutionProduct = "mdl_trx_institution_product"

	WrapMsgInsertInstitutionProduct          = WrapErrMsgPrefix + "InsertInstitutionProduct"
	WrapMsgUpdateTrxInstitutionProduct       = WrapErrMsgPrefix + "UpdateTrxInstitutionProduct"
	WrapMsgFindTrxInstitutionProductByParams = WrapErrMsgPrefix + "FindTrxInstitutionProductByParams"
)

func (c *Conn) InsertInstitutionProduct(ctx context.Context, product *model.TrxInstitutionProduct) (err error) {
	session := xorm.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}

	_, err = session.
		Table(TrxInstitutionProduct).
		InsertOne(product)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertInstitutionProduct)
		return
	}

	return
}

func (c *Conn) FindTrxInstitutionProductByParams(ctx context.Context, request model.TrxInstitutionProduct) (products []model.TrxInstitutionProduct, err error) {
	if request.IDMstInstitution == 0 {
		err = commonerr.SetNewNoInstitutionError()
		return
	}

	session := c.DB.SlaveDB.
		Context(ctx).
		Table(TrxInstitutionProduct)

	if request.ID > 0 {
		session.Where("mtip.id = ?", request.ID)
	}

	if len(request.Name) > 0 {
		session.Where(fmt.Sprintf("mtip.name ilike '%%%s%%'", request.Name))
	}

	if request.IDMstProduct.Valid {
		session.Where("mtip.id_mst_product = ?", request.IDMstProduct.Int64)
	}

	err = session.
		Alias("mtip").
		Where("id_mst_institution = ?", request.IDMstInstitution).
		Find(&products)
	if err != nil {
		err = errors.Wrap(err, WrapMsgFindTrxInstitutionProductByParams)
		return
	}

	return
}

func (c *Conn) UpdateTrxInstitutionProduct(ctx context.Context, request *model.TrxInstitutionProduct) (err error) {
	session := xorm.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}

	_, err = session.
		Table(TrxInstitutionProduct).
		ID(request.ID).
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateTrxInstitutionProduct)
		return
	}

	return
}
