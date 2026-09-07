package compensation

import (
	"context"
	"database/sql"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// DetectedAttribution is one clinical or map hit for a visit before merge.
// ClinicalRowID is the winning-type row id (procedure/diagnosis/map); 0 for anamnesa.
type DetectedAttribution struct {
	Type          model.ContributionSourceType
	StaffID       string
	Name          string
	ClinicalRowID int64
	ProcedureID   int64
	DiagnosisID   int64
	ProductID     sql.NullInt64
	Label         sql.NullString
}

// ContributorDB reads visit contribution sources. Clinical-table reads are
// allowed; writes stay off those tables.
type ContributorDB interface {
	DetectForVisit(ctx context.Context, institutionID, visitID int64) ([]DetectedAttribution, error)
}
