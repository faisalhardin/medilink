package visit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/faisalhardin/medilink/internal/entity/model"
	institutionRepo "github.com/faisalhardin/medilink/internal/entity/repo/institution"
	journeyDB "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	WrapErrMsgPrefix             = "VisitUC."
	WrapMsgInsertNewVisit        = WrapErrMsgPrefix + "InsertNewVisit"
	WrapMsgGetPatientVisits      = WrapErrMsgPrefix + "GetPatientVisits"
	WrapMsgUpdatePatientVisit    = WrapErrMsgPrefix + "UpdatePatientVisit"
	WrapMsgInsertVisitTouchpoint = WrapErrMsgPrefix + "InsertVisitTouchpoint"
	WrapMsgUpdateVisitTouchpoint = WrapErrMsgPrefix + "UpdateVisitTouchpoint"
	WrapMsgGetVisitTouchpoint    = WrapErrMsgPrefix + "GetVisitTouchpoint"
	WrapMsgInsertVisitProduct    = WrapErrMsgPrefix + "InsertVisitProduct"
	WrapMsgReduceProductStock    = WrapErrMsgPrefix + "ReduceProductStock"
)

type VisitUC struct {
	PatientDB       patientRepo.PatientDB
	InstitutionRepo institutionRepo.InstitutionDB
	Transaction     xorm.DBTransactionInterface
	JourneyDB       journeyDB.JourneyDB
}

func NewVisitUC(u *VisitUC) *VisitUC {
	return u
}

func (u *VisitUC) InsertNewVisit(ctx context.Context, req model.InsertNewVisitRequest) (err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	mstPatient, err := u.PatientDB.GetPatients(ctx, model.GetPatientParams{
		PatientUUIDs:  []string{req.PatientUUID},
		InstitutionID: userDetail.InstitutionID,
	})
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertNewVisit)
	}

	if len(mstPatient) == 0 {
		return commonerr.SetNewBadRequest("patient is not found", "no patient with given uuid")
	}

	req.IDMstPatient = mstPatient[0].ID
	err = u.PatientDB.RecordPatientVisit(ctx, &req.TrxPatientVisit)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertNewVisit)
	}

	return nil
}

func (u *VisitUC) GetPatientVisitDetail(ctx context.Context, req model.GetPatientVisitParams) (visitDetail model.GetPatientVisitDetailResponse, err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	req.IDMstInstitution = userDetail.InstitutionID

	errGroup, ctxWithCancel := errgroup.WithContext(ctx)
	var wg sync.WaitGroup
	wg.Add(2)

	errGroup.Go(func() error {
		defer wg.Done()
		visit, err := u.PatientDB.GetPatientVisits(ctxWithCancel, req)
		if err != nil {
			err = errors.Wrap(err, WrapMsgGetPatientVisits)
			return err
		}

		if len(visit) == 0 {
			return nil
		}
		visitDetail.TrxPatientVisit = visit[0].TrxPatientVisit
		visitDetail.MstPatient = visit[0].MstPatientInstitution

		return nil
	})

	errGroup.Go(func() error {
		defer wg.Done()
		dtlVisit, err := u.PatientDB.GetDtlPatientVisit(ctxWithCancel, model.DtlPatientVisit{
			IDTrxPatientVisit: req.IDPatientVisit,
		})
		if err != nil {
			err = errors.Wrap(err, WrapMsgGetPatientVisits)
			return err
		}
		visitDetail.DtlPatientVisit = dtlVisit
		return nil
	})

	errGroup.Go(func() error {
		wg.Wait()
		return nil
	})

	if err = errGroup.Wait(); err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisits)
		return
	}

	return
}

func (u *VisitUC) ListPatientVisits(ctx context.Context, req model.GetPatientVisitParams) (visitResponse []model.ListPatientVisitBoards, err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	req.IDMstInstitution = userDetail.InstitutionID

	visits, err := u.PatientDB.GetPatientVisits(ctx, req)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisits)
		return
	}

	for _, visit := range visits {
		visitResponse = append(visitResponse, model.ListPatientVisitBoards{
			ID:                          visit.TrxPatientVisit.ID,
			IDMstJourneyBoard:           visit.TrxPatientVisit.IDMstJourneyBoard,
			IDMstJourneyPoint:           visit.TrxPatientVisit.IDMstJourneyPoint,
			IDMstServicePoint:           visit.TrxPatientVisit.IDMstServicePoint,
			NameMstServicePoint:         visit.MstServicePoint.Name,
			Action:                      visit.TrxPatientVisit.Action,
			Status:                      visit.TrxPatientVisit.Status,
			Notes:                       visit.TrxPatientVisit.Notes,
			CreateTime:                  visit.TrxPatientVisit.CreateTime,
			Name:                        visit.MstPatientInstitution.Name,
			UUID:                        visit.MstPatientInstitution.UUID,
			Sex:                         visit.MstPatientInstitution.Sex,
			UpdateTimeMstJourneyPointID: visit.TrxPatientVisit.UpdateTimeMstJourneyPointID,
		})
	}

	return
}

func (u *VisitUC) UpdatePatientVisit(ctx context.Context, req model.UpdatePatientVisitRequest) (err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	req.IDMstInstitution = userDetail.InstitutionID

	_, err = u.PatientDB.UpdatePatientVisit(ctx, req)
	if err != nil {
		return errors.Wrap(err, WrapMsgUpdatePatientVisit)
	}

	return
}

func (u *VisitUC) ValidatePatientVisitExist(ctx context.Context, req model.DtlPatientVisitRequest) (err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	visit, err := u.PatientDB.GetPatientVisits(ctx, model.GetPatientVisitParams{
		IDPatientVisit:   req.IDTrxPatientVisit,
		IDMstInstitution: userDetail.InstitutionID,
	})
	if err != nil {
		return err
	}
	if len(visit) == 0 {
		err = commonerr.SetNewBadRequest("invalid", "no patient visit found")
		return
	}

	return nil
}

func (u *VisitUC) UpsertVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlPatientVisit model.DtlPatientVisit, err error) {
	if req.ID > 0 {
		return u.UpdateVisitTouchpoint(ctx, req)
	} else {
		return u.InsertVisitTouchpoint(ctx, req)
	}
}

func (u *VisitUC) InsertVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlPatientVisit model.DtlPatientVisit, err error) {
	if err = u.ValidatePatientVisitExist(ctx, req); err != nil {
		err = errors.Wrap(err, WrapMsgInsertVisitTouchpoint)
		return
	}

	user, _ := auth.GetUserDetailFromCtx(ctx)
	contributorsSlice := []string{user.Email}
	contributors, err := json.Marshal(contributorsSlice)
	if err != nil {
		return
	}

	mstJourneyPoint, err := u.JourneyDB.GetJourneyPoint(ctx, model.MstJourneyPoint{
		ID: req.IDMstJourneyPoint,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertVisitTouchpoint)
		return
	}

	dtlPatientVisit = model.DtlPatientVisit{
		JourneyPointName:  mstJourneyPoint.Name,
		Notes:             req.Notes,
		IDTrxPatientVisit: req.IDTrxPatientVisit,
		IDMstJourneyPoint: req.IDMstJourneyPoint,
		Contributors:      contributors,
		IDMstServicePoint: req.IDMstServicePoint.Int64,
	}

	err = u.PatientDB.InsertDtlPatientVisit(ctx, &dtlPatientVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertVisitTouchpoint)
		return
	}

	return
}

func (u *VisitUC) UpdateVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlPatientVisit model.DtlPatientVisit, err error) {

	if err = u.ValidatePatientVisitExist(ctx, req); err != nil {
		err = errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
		return
	}

	oldPatientVisit, err := u.PatientDB.GetDtlPatientVisitByID(ctx, req.ID)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
		return
	}

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	isNewContributor, err := oldPatientVisit.AddContributor(userDetail.Email)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
		return
	}

	newPatientVisit := &model.DtlPatientVisit{
		ID:    req.ID,
		Notes: req.Notes,
	}

	if isNewContributor {
		newPatientVisit.Contributors = oldPatientVisit.Contributors
	}

	err = u.PatientDB.UpdateDtlPatientVisit(ctx, newPatientVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
		return
	}
	return *newPatientVisit, nil
}

func (u *VisitUC) GetVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlVisit []model.DtlPatientVisit, err error) {

	if err = u.ValidatePatientVisitExist(ctx, req); err != nil {
		err = errors.Wrap(err, WrapMsgGetVisitTouchpoint)
		return
	}

	dtlVisit, err = u.PatientDB.GetDtlPatientVisit(ctx, model.DtlPatientVisit{
		ID:                req.ID,
		IDTrxPatientVisit: req.IDTrxPatientVisit,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetVisitTouchpoint)
		return
	}
	return
}

func (u *VisitUC) InsertVisitProduct(ctx context.Context, req model.InsertTrxVisitProductRequest) (err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	dtlPatientVisit, err := u.PatientDB.GetDtlPatientVisit(ctx, model.DtlPatientVisit{
		ID: req.IDDtlPatientVisit,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertVisitProduct)
		return
	}
	if len(dtlPatientVisit) == 0 {
		err = commonerr.SetNoVisitDetailError()
		return
	}

	requestedTrxInstitutionProducts := model.FindTrxInstitutionProductDBParams{
		IDMstInstitution: userDetail.InstitutionID,
	}
	mapRequestedProductIDToTrxInstitutionProducts := map[int64]model.PurchasedProduct{}
	for _, requestProduct := range req.Products {
		requestedTrxInstitutionProducts.ID = append(requestedTrxInstitutionProducts.ID, requestProduct.IDTrxInstitutionProduct)
		mapRequestedProductIDToTrxInstitutionProducts[requestProduct.IDTrxInstitutionProduct] = requestProduct
	}

	productItems, err := u.InstitutionRepo.FindTrxInstitutionProductJoinStockByParams(
		ctx, requestedTrxInstitutionProducts,
	)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertVisitProduct)
		return
	}

	if len(productItems) == 0 {
		err = commonerr.SetNewBadRequest("invalid", "item not found")
		return
	}
	if len(productItems) != len(req.Products) {
		err = commonerr.SetNewBadRequest("item not found", "at least one item is not found")
		return
	}

	session, _ := u.Transaction.Begin(ctx)
	defer u.Transaction.Finish(session, &err)

	for _, productItem := range productItems {

		quantity := mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].Quantity
		discountRate := mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].DiscountRate
		discountPrice := mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].DiscountPrice

		sumPrice := productItem.Price * float64(quantity)
		if discountPrice > 0 {
			sumPrice = sumPrice - discountPrice
		} else if discountRate > 0 {
			sumPrice = sumPrice * (1 - discountRate)
		}
		adjustedPrice := mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].AdjustedPrice
		if adjustedPrice > sumPrice {
			err = commonerr.SetNewBadRequest("invalid price", fmt.Sprintf("adjusted price must not exceed total price %v", sumPrice))
			return
		}
		err = u.PatientDB.InsertTrxVisitProduct(ctx, &model.TrxVisitProduct{
			IDTrxInstitutionProduct: productItem.ID,
			IDMstInstitution:        userDetail.InstitutionID,
			IDTrxPatientVisit:       dtlPatientVisit[0].IDTrxPatientVisit,
			IDDtlPatientVisit:       req.IDDtlPatientVisit,
			Quantity:                quantity,
			UnitType:                productItem.UnitType,
			Price:                   productItem.Price,
			DiscountRate:            discountRate,
			DiscountPrice:           discountPrice,
			TotalPrice:              sumPrice,
			AdjustedPrice:           adjustedPrice,
		})
		if err != nil {
			err = errors.Wrap(err, WrapMsgInsertVisitProduct)
			return
		}

		err = u.ReduceProductStock(ctx, ProductStockReducerRequest{
			ProductID: productItem.ID,
			Quantity:  int64(quantity),
		})
		if err != nil {
			err = errors.Wrap(err, WrapMsgInsertVisitProduct)
			return
		}

	}

	return nil
}

type ProductStockReducerRequest struct {
	ProductID int64
	Quantity  int64
}

func (u *VisitUC) ReduceProductStock(ctx context.Context, params ProductStockReducerRequest) (err error) {
	productStocks, err := u.InstitutionRepo.FindTrxInstitutionProductStockByParams(ctx, model.DtlInstitutionProductStock{
		IDTrxInstitutionProduct: params.ProductID,
	})
	if err != nil {
		return errors.Wrap(err, WrapMsgReduceProductStock)
	}

	if len(productStocks) == 0 {
		return commonerr.SetNewBadRequest("not found", "product not found")
	}

	productStock := productStocks[0]
	productStock.Quantity -= params.Quantity

	err = u.InstitutionRepo.UpdateDtlInstitutionProductStock(ctx, &productStock)
	if err != nil {
		return errors.Wrap(err, WrapMsgReduceProductStock)
	}

	return nil
}
