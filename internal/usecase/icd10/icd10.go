package icd10

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	icd10repo "github.com/faisalhardin/medilink/internal/entity/repo/icd10"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	wrapMsgSearch = "ICD10UC.Search"

	// defaultSearchLimit matches BACKEND_SPEC §4 (autocomplete returns up to 20).
	defaultSearchLimit = 20
)

// ICD10UC is the concrete implementation of the ICD-10 lookup usecase.
type ICD10UC struct {
	ICD10DB icd10repo.ICD10DB
}

func NewICD10UC(u *ICD10UC) *ICD10UC {
	return u
}

// Search returns up to req.Limit (default 20) ICD-10 entries matching req.Query.
// The route is JWT-protected — we still re-check the context here so a misconfigured
// router would fail loudly with 401 instead of leaking reference data.
func (u *ICD10UC) Search(ctx context.Context, req model.ICD10SearchRequest) ([]model.RefICD10, error) {
	if _, found := auth.GetUserDetailFromCtx(ctx); !found {
		return nil, commonerr.SetNewUnauthorizedAPICall()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	rows, err := u.ICD10DB.Search(ctx, req.Query, limit)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgSearch)
	}
	if rows == nil {
		rows = []model.RefICD10{}
	}
	return rows, nil
}
