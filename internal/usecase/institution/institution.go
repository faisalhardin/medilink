package institution

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	institutionRepo "github.com/faisalhardin/medilink/internal/entity/repo/institution"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"

	"github.com/pkg/errors"
)

var (
	WrapErrMsgPrefix                = "InstitutionUC."
	WrapMsgInsertInstitution        = WrapErrMsgPrefix + "InsertInstitution"
	WrapMsgFindInstitutionByParams  = WrapErrMsgPrefix + "FindInstitutionByParams"
	WrapMsgInserInstitutionProduct  = WrapErrMsgPrefix + "InserInstitutionProduct"
	WrapMsgUpdateInstitutionProduct = WrapErrMsgPrefix + "UpdateInstitutionProduct"
)

type InstitutionUC struct {
	InstitutionRepo institutionRepo.InstitutionDB
	Transaction     xorm.DBTransactionInterface
}

func NewInstitutionUC(uc *InstitutionUC) *InstitutionUC {
	return uc
}

func (uc *InstitutionUC) InsertInstitution(ctx context.Context, request model.CreateInstitutionRequest) (err error) {
	err = uc.InstitutionRepo.InsertNewInstitution(ctx,
		&model.Institution{
			Name: request.Name,
		})
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertInstitution)
		return err
	}

	return
}

func (uc *InstitutionUC) FindInstitutionByParams(ctx context.Context, params model.FindInstitutionParams) (result []model.Institution, err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	params.ID = userDetail.InstitutionID

	result, err = uc.InstitutionRepo.FindInstitutionByParams(ctx, params)
	if err != nil {
		err = errors.Wrap(err, WrapMsgFindInstitutionByParams)
		return
	}

	return
}
