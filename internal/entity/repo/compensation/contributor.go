package compensation

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// ErrContributorAlreadyAdded is a live unique (visit_id, staff_id) pair.
var ErrContributorAlreadyAdded = errors.New("contributor already added")

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

// PeriodStaffDetection is one contributing staff aggregated across visits in a payday period.
type PeriodStaffDetection struct {
	StaffID    string
	Name       string
	Roles      []string
	VisitCount int64
}

// ContributorDB reads visit contribution sources and writes mdl_map_visit_contributor.
// Clinical-table writes stay off this repo.
type ContributorDB interface {
	DetectForVisit(ctx context.Context, institutionID, visitID int64) ([]DetectedAttribution, error)
	DetectStaffForPeriod(ctx context.Context, institutionID int64, periodStart, periodEndExclusive time.Time) ([]PeriodStaffDetection, error)
	UpsertManualContributor(ctx context.Context, row model.MapVisitContributor) error
	DeleteManualContributor(ctx context.Context, institutionID, visitID int64, staffID string) (bool, error)
}
