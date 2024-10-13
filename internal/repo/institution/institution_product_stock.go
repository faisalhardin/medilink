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

func (c *Conn) FindTrxInstitutionProductJoinStockByParams(ctx context.Context, request model.FindTrxInstitutionProductDBParams) (products []model.GetInstitutionProductResponse, err error) {
	if request.IDMstInstitution == 0 {
		err = errors.Wrap(commonerr.SetNewNoInstitutionError(), WrapMsgFindTrxInstitutionProductJoinStockByParams)
		return
	}

	session := c.DB.SlaveDB.
		Table(TrxInstitutionProduct).
		Alias("mtip").
		Join(database.SQLInner, "mdl_dtl_institution_product_stock mdips", "(mtip.id = mdips.id_trx_institution_product and mtip.delete_time is null and mdips.delete_time is null)")

	if len(request.ID) > 0 {
		session.Where("mtip.id = ANY(?)", pq.Array(request.ID))
	}

	if len(request.Name) > 0 {
		session.Where(fmt.Sprintf("mtip.name ilike '%%%s%%'", request.Name))
	}

	if len(request.IDMstProduct) > 0 {
		session.Where("mtip.id_mst_product = ANY(?)", pq.Array(request.IDMstProduct))
	}

	err = session.
		Where("id_mst_institution = ?", request.IDMstInstitution).
		Select(`mtip.id, mtip.name, mtip.id_mst_product, mtip.price, 
		mtip.is_item, mtip.is_treatment, mdips.quantity, mdips.unit_type`).
		Find(&products)
	if err != nil {
		err = errors.Wrap(err, WrapMsgFindTrxInstitutionProductJoinStockByParams)
		return
	}

	return
}

func (c *Conn) UpdateDtlInstitutionProductStock(ctx context.Context, request *model.DtlInstitutionProductStock) (err error) {
	session := c.DB.MasterDB.Table(DtlInstitutionProductStock)

	_, err = session.
		Where("id_trx_institution_product = ? and delete_time is null", request.IDTrxInstitutionProduct).
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateDtlInstitutionProductStock)
		return
	}

	return
}
