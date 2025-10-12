package visit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	institutionRepo "github.com/faisalhardin/medilink/internal/entity/repo/institution"
	journeyDB "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	customtime "github.com/faisalhardin/medilink/pkg/type/time"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
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

const (
	defaultLimit = 5
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

	session, _ := u.Transaction.Begin(ctx)
	defer u.Transaction.Finish(session, &err)

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

	journeyPoint, err := u.JourneyDB.GetJourneyPoint(ctx, model.MstJourneyPoint{
		ShortID: req.JourneyPointShortID,
	})
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertNewVisit)
	}

	journeyBoard, err := u.JourneyDB.GetJourneyBoardByJourneyPoint(ctx, *journeyPoint)
	if err != nil && errors.Is(err, constant.ErrorRowNotFound) {
		return commonerr.SetNewBadRequest("journey point is not found", "no journey point with given id")
	}
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertNewVisit)
	}

	patientID := mstPatient[0].ID
	institutionID := userDetail.InstitutionID

	newTrxVisit := &model.TrxPatientVisit{
		IDMstPatient:                patientID,
		IDMstInstitution:            institutionID,
		IDMstJourneyPoint:           journeyPoint.ID,
		IDMstJourneyBoard:           journeyBoard.ID,
		UpdateTimeMstJourneyPointID: time.Now().Unix(),
	}

	err = u.PatientDB.RecordPatientVisit(ctx, newTrxVisit)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertNewVisit)
	}

	if req.Notes == nil {
		return
	}

	visitDetail := model.DtlPatientVisit{
		IDTrxPatientVisit: newTrxVisit.ID,
		Notes:             req.Notes,
		JourneyPointName:  journeyPoint.Name,
		IDMstJourneyPoint: journeyPoint.ID,
	}
	visitDetail.AddContributor(userDetail.Email)

	err = u.PatientDB.InsertDtlPatientVisit(ctx, &visitDetail)
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
		visitDetail.TrxPatientVisit.ShortIDMstJourneyPoint = visit[0].MstJourneyPoint.ShortID
		visitDetail.MstPatient = visit[0].MstPatientInstitution
		visitDetail.JourneyPoint = visit[0].MstJourneyPoint
		visitDetail.ServicePoint = visit[0].MstServicePoint
		visitDetail.IDMstJourneyBoard = visit[0].TrxPatientVisit.IDMstJourneyBoard

		return nil
	})

	errGroup.Go(func() error {
		defer wg.Done()
		dtlVisit, err := u.PatientDB.GetDtlPatientVisit(ctxWithCancel, model.GetDtlPatientVisitParams{
			IDsTrxPatientVisit: []int64{req.IDPatientVisit},
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

func (u *VisitUC) ListPatientVisitDetailed(ctx context.Context, req model.GetPatientVisitParams) (visitsDetails []model.GetPatientVisitDetailResponse, err error) {

	if req.CommonRequestPayload.Limit == 0 {
		req.CommonRequestPayload.Limit = defaultLimit
	}

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	req.IDMstInstitution = userDetail.InstitutionID

	visitsDetails = []model.GetPatientVisitDetailResponse{}

	visits, err := u.PatientDB.GetPatientVisits(ctx, req)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisits)
		return
	}

	if len(visits) == 0 {
		return
	}

	visitIDs := []int64{}
	for _, visit := range visits {
		visitIDs = append(visitIDs, visit.TrxPatientVisit.ID)
		visitsDetails = append(visitsDetails, model.GetPatientVisitDetailResponse{
			TrxPatientVisit: visit.TrxPatientVisit,
			MstPatient:      visit.MstPatientInstitution,
			JourneyPoint:    visit.MstJourneyPoint,
			ServicePoint:    visit.MstServicePoint,
		})
	}

	dtlVisits, err := u.PatientDB.GetDtlPatientVisit(ctx, model.GetDtlPatientVisitParams{
		IDsTrxPatientVisit: visitIDs,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisits)
		return
	}

	mapVisitIDtoDtlVisit := map[int64][]model.DtlPatientVisitWithShortID{}
	for _, detailVisit := range dtlVisits {
		if dtlVisit, ok := mapVisitIDtoDtlVisit[detailVisit.DtlPatientVisit.IDTrxPatientVisit]; ok {
			dtlVisit = append(dtlVisit, detailVisit)
			mapVisitIDtoDtlVisit[detailVisit.DtlPatientVisit.IDTrxPatientVisit] = dtlVisit
		} else {
			mapVisitIDtoDtlVisit[detailVisit.DtlPatientVisit.IDTrxPatientVisit] = []model.DtlPatientVisitWithShortID{
				detailVisit,
			}
		}
	}

	for i, response := range visitsDetails {
		response.DtlPatientVisit = mapVisitIDtoDtlVisit[response.ID]
		visitsDetails[i] = response
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
	if !req.ToTime.IsZero() {
		req.ToTime = customtime.Time{
			Time: time.Date(req.ToTime.Year(), req.ToTime.Month(), req.ToTime.Day(), 23, 59, 59, 0, req.ToTime.Location()),
		}
	}
	visits, err := u.PatientDB.GetPatientVisits(ctx, req)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisits)
		return
	}

	for _, visit := range visits {
		visitResponse = append(visitResponse, model.ListPatientVisitBoards{
			ID:                          visit.TrxPatientVisit.ID,
			IDMstJourneyBoard:           visit.TrxPatientVisit.IDMstJourneyBoard,
			ShortIDMstJourneyPoint:      visit.MstJourneyPoint.ShortID,
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

	if req.ShortIDMstJourneyPoint.Valid {
		journeyPoint, err := u.JourneyDB.GetJourneyPointByShortID(ctx, req.ShortIDMstJourneyPoint.String)
		if err != nil {
			return errors.Wrap(err, WrapMsgUpdatePatientVisit)
		}
		req.IDMstJourneyPoint = null.Int64From(journeyPoint.ID)
	}

	_, err = u.PatientDB.UpdatePatientVisit(ctx, req)
	if err != nil {
		return errors.Wrap(err, WrapMsgUpdatePatientVisit)
	}

	return
}

type ValidatePatientVisitExistRequest struct {
	IDTrxPatientVisit int64
}

func (u *VisitUC) ValidatePatientVisitExist(ctx context.Context, req ValidatePatientVisitExistRequest) (userDetail model.UserJWTPayload, err error) {
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
		return
	}
	if len(visit) == 0 {
		err = commonerr.SetNewBadRequest("invalid", "no patient visit found")
		return
	}

	return
}

func (u *VisitUC) UpsertVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlPatientVisit model.DtlPatientVisitWithShortID, err error) {
	if req.ID > 0 {
		return u.UpdateVisitTouchpoint(ctx, req)
	} else {
		return u.InsertVisitTouchpoint(ctx, req)
	}
}

func (u *VisitUC) InsertVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlPatientVisit model.DtlPatientVisitWithShortID, err error) {
	if _, err = u.ValidatePatientVisitExist(ctx, ValidatePatientVisitExistRequest{
		IDTrxPatientVisit: req.IDTrxPatientVisit,
	}); err != nil {
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
		ShortID: req.IDMstJourneyPoint,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertVisitTouchpoint)
		return
	}

	dtlPatientVisit = model.DtlPatientVisitWithShortID{
		DtlPatientVisit: model.DtlPatientVisit{
			JourneyPointName:  mstJourneyPoint.Name,
			Notes:             req.Notes,
			IDTrxPatientVisit: req.IDTrxPatientVisit,
			IDMstJourneyPoint: mstJourneyPoint.ID,
			Contributors:      contributors,
			IDMstServicePoint: req.IDMstServicePoint.Int64,
		},
		ShortIDMstJourneyPoint: mstJourneyPoint.ShortID,
	}

	err = u.PatientDB.InsertDtlPatientVisit(ctx, &dtlPatientVisit.DtlPatientVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertVisitTouchpoint)
		return
	}

	return
}

func (u *VisitUC) UpdateVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlPatientVisit model.DtlPatientVisitWithShortID, err error) {

	session, _ := u.Transaction.Begin(ctx)
	defer u.Transaction.Finish(session, &err)
	ctx = xorm.SetDBSession(ctx, session)

	if _, err = u.ValidatePatientVisitExist(ctx, ValidatePatientVisitExistRequest{
		IDTrxPatientVisit: req.IDTrxPatientVisit,
	}); err != nil {
		err = errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
		return
	}

	queryParams := model.GetDtlPatientVisitParams{}

	if req.ID > 0 {
		queryParams.IDs = []int64{req.ID}
	}
	if req.IDTrxPatientVisit > 0 {
		queryParams.IDsTrxPatientVisit = []int64{req.IDTrxPatientVisit}
	}
	if len(req.IDMstJourneyPoint) > 0 {
		queryParams.ShortIDsMstJourneyPoins = []string{req.IDMstJourneyPoint}
	}

	oldPatientVisits, err := u.PatientDB.GetDtlPatientVisit(ctx, queryParams)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
		return
	}

	if len(oldPatientVisits) == 0 {
		err = commonerr.SetNewBadRequest("invalid", "no patient visit found")
		return
	}

	oldPatientVisit := oldPatientVisits[0]

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	_, err = oldPatientVisit.AddContributor(userDetail.Email)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
		return
	}

	oldPatientVisit.Notes = req.Notes

	err = u.PatientDB.UpdateDtlPatientVisit(ctx, &oldPatientVisit.DtlPatientVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
		return
	}

	return oldPatientVisit, nil
}

func (u *VisitUC) GetVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlVisit []model.DtlPatientVisitWithShortID, err error) {

	if _, err = u.ValidatePatientVisitExist(ctx, ValidatePatientVisitExistRequest{
		IDTrxPatientVisit: req.IDTrxPatientVisit,
	}); err != nil {
		err = errors.Wrap(err, WrapMsgGetVisitTouchpoint)
		return
	}

	dtlVisit, err = u.PatientDB.GetDtlPatientVisit(ctx, model.GetDtlPatientVisitParams{
		IDs:                []int64{req.ID},
		IDsTrxPatientVisit: []int64{req.IDTrxPatientVisit},
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetVisitTouchpoint)
		return
	}
	return
}

// InsertVisitProduct processes the insertion of products into a patient visit.
// This function handles the complete workflow of adding products to a visit including:
// - User authentication and authorization
// - Visit detail validation
// - Product availability and stock validation
// - Price calculations with discounts
// - Database transaction management
// - Stock reduction
//
// Parameters:
// - ctx: Context for request lifecycle and cancellation
// - req: InsertTrxVisitProductRequest containing visit ID and list of products to add
//
// Returns:
// - error: nil on success, or wrapped error with context on failure
func (u *VisitUC) InsertVisitProduct(ctx context.Context, req model.InsertTrxVisitProductRequest) (err error) {

	// === USER AUTHENTICATION SECTION ===
	// Extract user details from context to ensure the request is authenticated
	// and get the institution ID for data isolation
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	// === VISIT DETAIL VALIDATION SECTION ===
	// Validate that the specified patient visit detail exists
	// This ensures we're adding products to a valid visit
	dtlPatientVisit, err := u.PatientDB.GetDtlPatientVisit(ctx, model.GetDtlPatientVisitParams{
		IDs: []int64{req.IDDtlPatientVisit},
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertVisitProduct)
		return
	}
	if len(dtlPatientVisit) == 0 {
		err = commonerr.SetNoVisitDetailError()
		return
	}

	// === PRODUCT REQUEST PREPARATION SECTION ===
	// Prepare parameters to fetch institution products and create a mapping
	// for quick lookup of requested product details
	requestedTrxInstitutionProducts := model.FindTrxInstitutionProductParams{
		IDMstInstitution: userDetail.InstitutionID,
	}
	// Create a map to quickly access requested product details by product ID
	mapRequestedProductIDToTrxInstitutionProducts := map[int64]model.PurchasedProduct{}
	for _, requestProduct := range req.Products {
		// Collect all requested product IDs for batch fetching
		requestedTrxInstitutionProducts.IDs = append(requestedTrxInstitutionProducts.IDs, requestProduct.IDTrxInstitutionProduct)
		// Map product ID to its request details for later reference
		mapRequestedProductIDToTrxInstitutionProducts[requestProduct.IDTrxInstitutionProduct] = requestProduct
	}

	// === PRODUCT VALIDATION SECTION ===
	// Fetch product details with current stock information from the database
	productItems, err := u.InstitutionRepo.FindTrxInstitutionProductJoinStockByParams(
		ctx, requestedTrxInstitutionProducts,
	)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertVisitProduct)
		return
	}

	// Validate that all requested products exist
	if len(productItems) == 0 {
		err = commonerr.SetNewBadRequest("invalid", "item not found")
		return
	}
	// Ensure all requested products were found (no missing products)
	if len(productItems) != len(req.Products) {
		err = commonerr.SetNewBadRequest("item not found", "at least one item is not found")
		return
	}

	// === TRANSACTION PROCESSING SECTION ===
	// Begin database transaction to ensure data consistency
	// All operations will be rolled back if any step fails
	session, _ := u.Transaction.Begin(ctx)
	defer u.Transaction.Finish(session, &err)

	// Process each product item in the request
	for _, productItem := range productItems {

		// Validate stock availability before processing
		if productItem.Quantity < int64(mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].Quantity) {
			err = commonerr.SetNewBadRequest("invalid", "purchase quantity exceeds stock")
			return
		}

		// Extract product details from the request mapping
		quantity := mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].Quantity
		discountRate := mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].DiscountRate
		discountPrice := mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].DiscountPrice

		// Calculate total price with discount logic
		sumPrice := productItem.Price * float64(quantity)
		if discountPrice > 0 {
			// Apply fixed discount amount if specified
			sumPrice = sumPrice - discountPrice
		} else if discountRate > 0 {
			// Apply percentage discount if specified
			sumPrice = sumPrice * (1 - discountRate)
		}

		// Validate adjusted price doesn't exceed calculated total
		adjustedPrice := mapRequestedProductIDToTrxInstitutionProducts[productItem.ID].AdjustedPrice
		if adjustedPrice > sumPrice {
			err = commonerr.SetNewBadRequest("invalid price", fmt.Sprintf("adjusted price must not exceed total price %v", sumPrice))
			return
		}

		// Insert the visit product record into the database
		err = u.PatientDB.InsertTrxVisitProduct(ctx, &model.TrxVisitProduct{
			IDTrxInstitutionProduct: productItem.ID,
			IDMstInstitution:        userDetail.InstitutionID,
			IDTrxPatientVisit:       dtlPatientVisit[0].DtlPatientVisit.IDTrxPatientVisit,
			IDDtlPatientVisit:       req.IDTrxPatientVisit,
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

		// Reduce the product stock by the purchased quantity
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

func (u *VisitUC) UpsertVisitProduct(ctx context.Context, req model.UpsertTrxVisitProductRequest) (err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	// START: fetch institution product
	mappedProductInstitution, err := u.getMappedInstitutionProducts(ctx, userDetail.InstitutionID, req.Products)
	if err != nil {
		return
	}
	// END:fetch institution product

	// START: fetch existing product
	mappedProductVisit, err := u.getMappedOrderedProduct(ctx, model.TrxVisitProduct{
		ID:               req.IDTrxPatientVisit,
		IDMstInstitution: userDetail.InstitutionID,
	})
	if err != nil {
		return
	}
	// END: fetch existing product

	// create mapping product id to requested product id
	// compare

	for _, requestedProduct := range req.Products {
		productStock := mappedProductInstitution[requestedProduct.IDTrxInstitutionProduct]

		orderedProduct, found := mappedProductVisit[requestedProduct.IDTrxInstitutionProduct]
		// if not exist then buy anew
		if !found {
			err = u.orderProduct(ctx, model.TrxVisitProduct{
				IDTrxInstitutionProduct: requestedProduct.IDTrxInstitutionProduct,
				IDMstInstitution:        userDetail.InstitutionID,
				IDTrxPatientVisit:       req.IDTrxPatientVisit,
				Name:                    productStock.Name,
				IDDtlPatientVisit:       req.IDDtlPatientVisit,
				UnitType:                productStock.UnitType,
				Price:                   productStock.Price,
				DiscountRate:            requestedProduct.DiscountRate,
				DiscountPrice:           requestedProduct.DiscountPrice,
				TotalPrice:              float64(requestedProduct.TotalPrice),
				AdjustedPrice:           requestedProduct.AdjustedPrice,
			},
				productStock,
				requestedProduct)
			if err != nil {
				return
			}
			continue
		}

		// if requested quantity != existing quantity => reduce/increas from stock and add/substract to visit product
		if requestedProduct.Quantity != orderedProduct.Quantity {

			err = u.orderProduct(
				ctx,
				orderedProduct,
				productStock,
				requestedProduct)
			if err != nil {
				return
			}

		}
		// if requested quantity == existing quantity => do nothing

		// delete product from mapping
		delete(mappedProductVisit, requestedProduct.IDTrxInstitutionProduct)
	}

	// if exist product from mapping => add product quantity back to stock then delete its visit product record
	for _, remainingProduct := range mappedProductVisit {
		// productStock := mappedProductInstitution[remainingProduct.IDTrxInstitutionProduct]
		err = u.voidOrder(ctx, remainingProduct)
		if err != nil {
			return
		}

	}

	return nil
}

func (u *VisitUC) voidOrder(ctx context.Context,
	existingProduct model.TrxVisitProduct,
) (err error) {
	err = u.InstitutionRepo.RestockDtlInstitutionProductStock(ctx, &model.DtlInstitutionProductStock{
		IDTrxInstitutionProduct: existingProduct.IDTrxInstitutionProduct,
		Quantity:                int64(existingProduct.Quantity),
	})
	if err != nil {
		return errors.Wrap(err, "usecase.voidOrder")
	}

	err = u.PatientDB.DeleteTrxVisitProduct(ctx, &existingProduct)
	if err != nil {
		return
	}

	return nil
}

func (u *VisitUC) orderProduct(ctx context.Context,
	existingProduct model.TrxVisitProduct,
	productStock model.GetInstitutionProductResponse,
	productRequest model.PurchasedProduct) (err error) {

	quantityDifference := existingProduct.Quantity - productRequest.Quantity

	existingStock := model.DtlInstitutionProductStock{
		IDTrxInstitutionProduct: productStock.ID,
		Quantity:                productStock.Quantity + int64(quantityDifference), // if less than zero return error below
	}

	// if existing quantity > requested quantity => stock replenished
	// if existing qunatity < reuqested quantity => stock reduced
	if existingStock.Quantity < 0 && productStock.IsItem {
		return commonerr.SetNewBadRequest("stock issue", fmt.Sprintf("product %s stock is not enough", productStock.Name))
	}

	if productStock.IsItem {
		err = u.InstitutionRepo.UpdateDtlInstitutionProductStock(ctx, &existingStock)
		if err != nil {
			return errors.Wrap(err, "orderNewProductForVisit")
		}
	}

	existingProduct.Quantity = productRequest.Quantity
	existingProduct.AdjustedPrice = productRequest.AdjustedPrice
	existingProduct.DiscountPrice = productRequest.DiscountPrice
	existingProduct.DiscountRate = productRequest.DiscountRate
	decimalPrice := decimal.NewFromFloat(productStock.Price)

	existingProduct.TotalPrice = decimal.NewFromInt(int64(existingProduct.Quantity)).Mul(decimalPrice).InexactFloat64()
	err = u.PatientDB.UpsertTrxVisitProduct(ctx, &existingProduct)
	if err != nil {
		return errors.Wrap(err, "orderNewProductForVisit")
	}

	return
}

func (u *VisitUC) getMappedInstitutionProducts(
	ctx context.Context,
	institutionID int64,
	requestedProducts []model.PurchasedProduct,
) (map[int64]model.GetInstitutionProductResponse, error) {

	requestedTrxInstitutionProducts := model.FindTrxInstitutionProductParams{
		IDMstInstitution: institutionID,
	}
	// Create a map to quickly access requested product details by product ID
	mapRequestedProductIDToTrxInstitutionProducts := map[int64]model.PurchasedProduct{}
	for _, requestProduct := range requestedProducts {
		// Collect all requested product IDs for batch fetching
		requestedTrxInstitutionProducts.IDs = append(requestedTrxInstitutionProducts.IDs, requestProduct.IDTrxInstitutionProduct)
		// Map product ID to its request details for later reference
		mapRequestedProductIDToTrxInstitutionProducts[requestProduct.IDTrxInstitutionProduct] = requestProduct
	}

	productItems, err := u.InstitutionRepo.FindTrxInstitutionProductJoinStockByParams(
		ctx, requestedTrxInstitutionProducts,
	)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgInsertVisitProduct)
	}

	// Validate that all requested products exist
	if len(productItems) == 0 {
		return nil, commonerr.SetNewBadRequest("invalid", "item not found")
	}
	// Ensure all requested products were found (no missing products)
	if len(productItems) != len(requestedProducts) {
		return nil, commonerr.SetNewBadRequest("item not found", "at least one item is not found")
	}

	// Create the final map of product ID to TrxInstitutionProductJoinStock
	mappedProductItems := make(map[int64]model.GetInstitutionProductResponse)
	for _, item := range productItems {
		mappedProductItems[item.ID] = item
	}

	return mappedProductItems, nil
}

func (u *VisitUC) getMappedOrderedProduct(ctx context.Context, trxVisit model.TrxVisitProduct) (mappedProductVisitByProductID map[int64]model.TrxVisitProduct, err error) {
	orderedProducts, err := u.PatientDB.GetTrxVisitProduct(ctx, model.GetVisitProductRequest{
		VisitID:       trxVisit.ID,
		InstitutionID: trxVisit.IDMstInstitution,
	})
	if err != nil {
		return
	}

	mappedProductVisitByProductID = make(map[int64]model.TrxVisitProduct)
	for _, orderedProduct := range orderedProducts {
		mappedProductVisitByProductID[orderedProduct.IDTrxInstitutionProduct] = orderedProduct
	}

	return
}

func (u *VisitUC) UpdateVisitProduct(ctx context.Context, req model.InsertTrxVisitProductRequest) (err error) {
	_, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	err = u.PatientDB.UpdateDtlPatientVisit(ctx, &model.DtlPatientVisit{
		ID: req.IDDtlPatientVisit,
	})

	return
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

func (u *VisitUC) ListVisitProducts(ctx context.Context, params model.GetVisitProductRequest) (products []model.TrxVisitProduct, err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	params.InstitutionID = userDetail.InstitutionID
	return u.PatientDB.GetTrxVisitProduct(ctx, params)
}

func (u *VisitUC) ArchivePatientVisit(ctx context.Context, req model.ArchivePatientVisitRequest) (err error) {

	_, err = u.ValidatePatientVisitExist(ctx, ValidatePatientVisitExistRequest{
		IDTrxPatientVisit: req.ID,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
		return
	}

	err = u.PatientDB.DeletePatientVisit(ctx, &model.TrxPatientVisit{
		ID: req.ID,
	})
	if err != nil {
		return errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
	}

	return nil
}
