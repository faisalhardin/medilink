package institution

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/constant/database"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	DtlInstitutionProductStock = "mdl_dtl_institution_product_stock"

	WrapMsgFindTrxInstitutionProductJoinStockByParams = "FindTrxInstitutionProductJoinStockByParams"
	WrapMsgUpdateDtlInstitutionProductStock           = "UpdateDtlInstitutionProductStock"
)

func (c *Conn) InsertInstitutionProductStock(ctx context.Context, product *model.DtlInstitutionProductStock) (err error) {
	session := c.DB.MasterDB.Table(DtlInstitutionProductStock)

	_, err = session.InsertOne(product)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertInstitutionProduct)
		return
	}

	return
}

func (c *Conn) FindTrxInstitutionProductStockByParams(ctx context.Context, request model.DtlInstitutionProductStock) (stock []model.DtlInstitutionProductStock, err error) {

	session := c.DB.SlaveDB.Table(DtlInstitutionProductStock).Alias("mdips")

	err = session.
		Where("id_trx_institution_product = ?", request.IDTrxInstitutionProduct).
		Find(&stock)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetInstitution)
		return
	}

	return
}

func (c *Conn) FindTrxInstitutionProductJoinStockByParams(ctx context.Context, request model.FindTrxInstitutionProductParams) (products []model.GetInstitutionProductResponse, err error) {
	if request.IDMstInstitution == 0 {
		err = errors.Wrap(commonerr.SetNewNoInstitutionError(), WrapMsgFindTrxInstitutionProductJoinStockByParams)
		return
	}
	if request.Page > 0 {
		request.Offset = request.Limit * (request.Page - 1)
	}
	if request.Limit == 0 {
		request.Limit = 30
	}

	session := c.DB.SlaveDB.
		Table(TrxInstitutionProduct).
		Alias("mtip").
		Join(database.SQLInner, "mdl_dtl_institution_product_stock mdips", "(mtip.id = mdips.id_trx_institution_product and mtip.delete_time is null and mdips.delete_time is null)")

	if len(request.IDs) > 0 {
		session.Where("mtip.id = ANY(?)", pq.Array(request.IDs))
	}

	if len(request.Name) > 0 {
		session.Where(fmt.Sprintf("mtip.name ilike '%%%s%%'", request.Name))
	}
	if request.IsItem {
		session.Where("mtip.is_item = ?", request.IsItem)
	}
	if request.IsTreatment {
		session.Where("mtip.is_treatment = ?", request.IsTreatment)
	}

	if len(request.IDMstProduct) > 0 {
		session.Where("mtip.id_mst_product = ANY(?)", pq.Array(request.IDMstProduct))
	}

	if request.Limit > 0 {
		session.Limit(request.Limit, request.Offset)
	}

	err = session.
		Where("id_mst_institution = ?", request.IDMstInstitution).
		Select(`mtip.id, mtip.name, mtip.id_mst_product, mtip.price, 
		mtip.is_item, mtip.is_treatment, mdips.quantity, mdips.unit_type`).
		OrderBy("mtip.id DESC").
		Find(&products)
	if err != nil {
		err = errors.Wrap(err, WrapMsgFindTrxInstitutionProductJoinStockByParams)
		return
	}

	return
}

func (c *Conn) UpdateDtlInstitutionProductStock(ctx context.Context, request *model.DtlInstitutionProductStock) (err error) {
	session := c.DB.MasterDB.Table(DtlInstitutionProductStock)

	if request.ID == 0 && request.IDTrxInstitutionProduct == 0 {
		err = errors.New("invalid parameter: id or id_trx_institution_product must be defined")
		err = errors.Wrap(err, WrapMsgUpdateDtlInstitutionProductStock)
		return
	}

	if request.ID > 0 {
		session.Where("id = ?", request.ID)
	}

	if request.IDTrxInstitutionProduct > 0 {
		session.Where("id_trx_institution_product = ? ", request.IDTrxInstitutionProduct)
	}

	_, err = session.
		Cols("quantity").
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateDtlInstitutionProductStock)
		return
	}

	return
}

func (c *Conn) RestockDtlInstitutionProductStock(ctx context.Context, request *model.DtlInstitutionProductStock) (err error) {
	session := c.DB.MasterDB.Table(DtlInstitutionProductStock)

	if request.ID == 0 && request.IDTrxInstitutionProduct == 0 {
		err = errors.New("invalid parameter: id or id_trx_institution_product must be defined")
		err = errors.Wrap(err, WrapMsgUpdateDtlInstitutionProductStock)
		return
	}

	if request.ID > 0 {
		session.Where("id = ?", request.ID)
	}

	if request.IDTrxInstitutionProduct > 0 {
		session.Where("id_trx_institution_product = ? ", request.IDTrxInstitutionProduct)
	}

	_, err = session.Incr("quantity", request.Quantity).Update(&model.DtlInstitutionProductStock{})
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateDtlInstitutionProductStock)
		return
	}

	return
}
