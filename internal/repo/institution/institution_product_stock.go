package institution

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/constant/database"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/pkg/errors"
)

const (
	TrxInstitutionProductStock = "mdl_trx_institution_product_stock"

	WrapMsgFindTrxInstitutionProductJoinStockByParams = "FindTrxInstitutionProductJoinStockByParams"
	WrapMsgUpdateTrxInstitutionProductStock           = "UpdateTrxInstitutionProductStock"
)

func (c *Conn) InsertInstitutionProductStock(ctx context.Context, product *model.TrxInstitutionProductStock) (err error) {
	session := c.DB.MasterDB.Table(TrxInstitutionProductStock)

	_, err = session.InsertOne(product)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertInstitutionProduct)
		return
	}

	return
}

func (c *Conn) FindTrxInstitutionProductStockByParams(ctx context.Context, request model.TrxInstitutionProductStock) (stock []model.TrxInstitutionProductStock, err error) {

	session := c.DB.SlaveDB.Table(TrxInstitutionProductStock).Alias("mtips")

	err = session.
		Where("id_mst_institution_product = ?", request.IDMstInstitutionProduct).
		Find(&stock)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetInstitution)
		return
	}

	return
}

func (c *Conn) FindTrxInstitutionProductJoinStockByParams(ctx context.Context, request model.TrxInstitutionProduct) (products []model.TrxInstitutionProductJoinStock, err error) {
	if request.IDMstInstitution == 0 {
		err = errors.Wrap(commonerr.SetNewNoInstitutionError(), WrapMsgFindTrxInstitutionProductJoinStockByParams)
		return
	}

	session := c.DB.SlaveDB.
		Table(TrxInstitutionProduct).
		Alias("mtip").
		Join(database.SQLInner, "mdl_trx_institution_product_stock mtips", "mtip.id = mtips.id_mst_institution_product and mtip.delete_time is null and and mtips.delete_time is null")

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
		err = errors.Wrap(err, WrapMsgFindTrxInstitutionProductJoinStockByParams)
		return
	}

	return
}

func (c *Conn) UpdateTrxInstitutionProductStock(ctx context.Context, request *model.TrxInstitutionProductStock) (err error) {
	session := c.DB.MasterDB.Table(TrxInstitutionProductStock)

	_, err = session.
		Where("id_mst_institution_product = ? and delete_time is null", request.IDMstInstitutionProduct).
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateTrxInstitutionProduct)
		return
	}

	return
}
