package institution

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/pkg/errors"
)

const (
	TrxInstitutionProduct = "mdl_trx_institution_product"

	WrapMsgInsertInstitutionProduct          = WrapErrMsgPrefix + "InsertInstitutionProduct"
	WrapMsgUpdateTrxInstitutionProduct       = WrapErrMsgPrefix + "UpdateTrxInstitutionProduct"
	WrapMsgFindTrxInstitutionProductByParams = WrapErrMsgPrefix + "FindTrxInstitutionProductByParams"
)

func (c *Conn) InsertInstitutionProduct(ctx context.Context, product *model.TrxInstitutionProduct) (err error) {
	session := c.DB.MasterDB.Table(TrxInstitutionProduct)

	_, err = session.InsertOne(product)
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

	session := c.DB.SlaveDB.Table(TrxInstitutionProduct).Alias("mtip")

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
		Where("id_mst_institution = ?", request.IDMstInstitution).
		Find(&products)
	if err != nil {
		err = errors.Wrap(err, WrapMsgFindTrxInstitutionProductByParams)
		return
	}

	return
}

func (c *Conn) UpdateTrxInstitutionProduct(ctx context.Context, request *model.TrxInstitutionProduct) (err error) {
	session := c.DB.MasterDB.Table(TrxInstitutionProduct)

	_, err = session.
		ID(request.ID).
		Cols("name", "id_mst_product", "price").
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateTrxInstitutionProduct)
		return
	}

	return
}
