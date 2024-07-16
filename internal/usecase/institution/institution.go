package institution

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	institutionRepo "github.com/faisalhardin/medilink/internal/entity/repo/institution"

	"github.com/pkg/errors"
)

var (
	WrapErrMsgPrefix               = "InstitutionUC."
	WrapMsgInsertInstitution       = WrapErrMsgPrefix + "InsertInstitution"
	WrapMsgFindInstitutionByParams = WrapErrMsgPrefix + "FindInstitutionByParams"
)

type InstitutionUC struct {
	InstitutionRepo institutionRepo.InstitutionDB
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
	result, err = uc.InstitutionRepo.FindInstitutionByParams(ctx, params)
	if err != nil {
		err = errors.Wrap(err, WrapMsgFindInstitutionByParams)
		return
	}

	return
}
