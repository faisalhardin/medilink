package visit

import (
	"context"
	"sync"

	"github.com/faisalhardin/medilink/internal/entity/model"
	institutionRepo "github.com/faisalhardin/medilink/internal/entity/repo/institution"
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
)

type VisitUC struct {
	PatientDB       patientRepo.PatientDB
	InstitutionRepo institutionRepo.InstitutionDB
	Transaction     xorm.DBTransactionInterface
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
		visitDetail.TrxPatientVisit = visit[0]
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

func (u *VisitUC) ListPatientVisits(ctx context.Context, req model.GetPatientVisitParams) (visits []model.TrxPatientVisit, err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	req.IDMstInstitution = userDetail.InstitutionID

	visits, err = u.PatientDB.GetPatientVisits(ctx, req)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisits)
		return
	}

	return
}

func (u *VisitUC) UpdatePatientVisit(ctx context.Context, req model.UpdatePatientVisitRequest) (err error) {
	err = u.PatientDB.UpdatePatientVisit(ctx, req.TrxPatientVisit)
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

func (u *VisitUC) InsertVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (err error) {

	if err = u.ValidatePatientVisitExist(ctx, req); err != nil {
		return errors.Wrap(err, WrapMsgInsertVisitTouchpoint)
	}

	err = u.PatientDB.InsertDtlPatientVisit(ctx, &model.DtlPatientVisit{
		IDTrxPatientVisit:  req.IDTrxPatientVisit,
		TouchpointName:     req.TouchpointName,
		TouchpointCategory: req.TouchpointCategory,
		Notes:              req.Notes,
	})
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertVisitTouchpoint)
	}
	return
}

func (u *VisitUC) UpdateVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (err error) {

	if err = u.ValidatePatientVisitExist(ctx, req); err != nil {
		return errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
	}

	err = u.PatientDB.UpdateDtlPatientVisit(ctx, &model.DtlPatientVisit{
		ID:                 req.ID,
		TouchpointName:     req.TouchpointName,
		TouchpointCategory: req.TouchpointCategory,
		Notes:              req.Notes,
	})
	if err != nil {
		return errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
	}
	return
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
		err = u.PatientDB.InsertTrxVisitProduct(ctx, &model.TrxVisitProduct{
			IDTrxInstitutionProduct: productItem.ID,
			IDMstInstitution:        userDetail.InstitutionID,
			IDTrxPatientVisit:       dtlPatientVisit[0].IDTrxPatientVisit,
			IDDtlPatientVisit:       req.IDDtlPatientVisit,
			Quantity:                mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].Quantity, //int(productItem.Quantity),
			UnitType:                productItem.UnitType,
			Price:                   productItem.Price,
			DiscountRate:            mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].DiscountRate,
			DiscountPrice:           mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].DiscountPrice,
			TotalPrice:              productItem.Price * float64(mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].Quantity),
			AdjustedPrice:           mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].AdjustedPrice,
		})
		if err != nil {
			err = errors.Wrap(err, WrapMsgInsertVisitProduct)
			return
		}
	}

	return nil
}
