package institution

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	TrxInstitutionProduct = "mdl_trx_institution_product"

	WrapMsgInsertInstitutionProduct          = WrapErrMsgPrefix + "InsertInstitutionProduct"
	WrapMsgUpdateTrxInstitutionProduct       = WrapErrMsgPrefix + "UpdateTrxInstitutionProduct"
	WrapMsgFindTrxInstitutionProductByParams = WrapErrMsgPrefix + "FindTrxInstitutionProductByParams"
	WrapMsgGetProductStatistics              = WrapErrMsgPrefix + "GetProductStatistics"
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

func (c *Conn) FindTrxInstitutionProductByParams(ctx context.Context, request model.FindTrxInstitutionProductParams) (products []model.TrxInstitutionProduct, err error) {
	if request.IDMstInstitution == 0 {
		err = commonerr.SetNewNoInstitutionError()
		return
	}

	session := c.DB.SlaveDB.
		Context(ctx).
		Table(TrxInstitutionProduct)

	if len(request.IDs) > 0 {
		session.Where("mtip.id = any(?)", pq.Array(request.IDs))
	}
	if request.IsItem {
		session.Where("mtip.is_item = ?", request.IsItem)
	}
	if request.IsTreatment {
		session.Where("mtip.is_treatment = ?", request.IsTreatment)
	}

	if len(request.Name) > 0 {
		session.Where(fmt.Sprintf("mtip.name ilike '%%%s%%'", request.Name))
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

func (c *Conn) UpdateTrxInstitutionProduct(ctx context.Context, request *model.UpdateInstitutionProductRequest) (resp model.TrxInstitutionProduct, err error) {
	session := xorm.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}

	trxProduct := &model.TrxInstitutionProduct{
		ID:           request.ID,
		Name:         request.Name,
		IDMstProduct: request.IDMstProduct,
		Price:        request.Price,
		IsItem:       request.IsItem.Bool,
		IsTreatment:  request.IsTreatment.Bool,
	}

	if request.IsItem.Valid {
		session.UseBool("is_item")
	}
	if request.IsTreatment.Valid {
		session.UseBool("is_treatment")
	}

	_, err = session.
		Table(TrxInstitutionProduct).
		ID(request.ID).
		Update(trxProduct)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateTrxInstitutionProduct)
		return
	}

	return
}

func (c *Conn) GetProductStatistics(ctx context.Context, query model.ProductStatisticsQuery) (rows []model.ProductStatisticsRow, err error) {
	if query.IDMstInstitution == 0 {
		err = commonerr.SetNewNoInstitutionError()
		return
	}

	_, offsetSec := query.StartTime.Zone()

	sql := `
		SELECT
			date_trunc('` + query.Granularity + `', (vp.create_time AT TIME ZONE 'UTC') + (? * interval '1 second')) AS period_start,
			vp.id_trx_institution_product,
			vp.name,
			COALESCE(SUM(vp.quantity), 0)::bigint AS total_quantity,
			COALESCE(SUM(vp.total_price), 0) AS total_revenue,
			CASE
				WHEN COALESCE(SUM(vp.quantity), 0) > 0
				THEN COALESCE(SUM(vp.total_price), 0) / SUM(vp.quantity)
				ELSE 0
			END AS avg_unit_price
		FROM ` + model.TrxVisitProductTableName + ` vp
		WHERE vp.id_mst_institution = ?
		  AND vp.create_time >= ?
		  AND vp.create_time <= ?
		  AND vp.delete_time IS NULL
	`

	args := []interface{}{
		offsetSec,
		query.IDMstInstitution,
		query.StartTime.UTC(),
		query.EndTime.UTC(),
	}

	if query.IDTrxInstitutionProduct > 0 {
		sql += ` AND vp.id_trx_institution_product = ?`
		args = append(args, query.IDTrxInstitutionProduct)
	}

	sql += `
		GROUP BY period_start, vp.id_trx_institution_product, vp.name
		ORDER BY period_start ASC, vp.name ASC
	`

	err = c.DB.SlaveDB.Context(ctx).SQL(sql, args...).Find(&rows)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetProductStatistics)
		return
	}

	return
}
