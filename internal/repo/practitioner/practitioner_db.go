package practitioner

import (
	"context"
	"strings"

	"github.com/faisalhardin/medilink/internal/entity/model"
	practitionerrepo "github.com/faisalhardin/medilink/internal/entity/repo/practitioner"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix       = "PractitionerDB."
	WrapMsgSearchDoctors   = WrapErrMsgPrefix + "SearchDoctors"
	WrapMsgSearchNurses    = WrapErrMsgPrefix + "SearchNurses"
	WrapMsgMissingDoctorID = WrapErrMsgPrefix + "MissingDoctorIDs"

	defaultSearchLimit = 10
	maxSearchLimit     = 50
)

type Conn struct {
	DB *xormlib.DBConnect
}

// NewPractitionerDB returns a PractitionerDB implementation bound to the given xorm connection.
func NewPractitionerDB(db *xormlib.DBConnect) practitionerrepo.PractitionerDB {
	return &Conn{DB: db}
}

// SearchDoctors runs prefix ILIKE + tsvector match scoped by institution.
// Rows linked to a staff account are surfaced first so internal practitioners
// show up before externally-registered ones.
func (c *Conn) SearchDoctors(ctx context.Context, institutionID int64, q string, limit int) ([]model.DoctorSearchResult, error) {
	q = strings.TrimSpace(q)
	limit = clampLimit(limit)

	const sql = `
		SELECT id, staff_uuid, name, sip_number, specialization
		FROM mdl_mst_doctor
		WHERE institution_id = ?
		  AND active = TRUE
		  AND (
		    ? = ''
		    OR name ILIKE ? || '%'
		    OR to_tsvector('simple', name) @@ plainto_tsquery('simple', ?)
		  )
		ORDER BY (staff_uuid IS NULL), name
		LIMIT ?
	`

	var rows []model.DoctorSearchResult
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, institutionID, q, q, q, limit).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgSearchDoctors)
	}
	return rows, nil
}

// SearchNurses mirrors SearchDoctors with an optional role filter.
func (c *Conn) SearchNurses(ctx context.Context, institutionID int64, role *string, q string, limit int) ([]model.NurseSearchResult, error) {
	q = strings.TrimSpace(q)
	limit = clampLimit(limit)

	args := []interface{}{institutionID, q, q, q}
	roleClause := ""
	if role != nil && *role != "" {
		roleClause = " AND role = ?"
		args = append(args, *role)
	}
	args = append(args, limit)

	sql := `
		SELECT id, staff_uuid, name, sip_number, role
		FROM mdl_mst_nurse
		WHERE institution_id = ?
		  AND active = TRUE
		  AND (
		    ? = ''
		    OR name ILIKE ? || '%'
		    OR to_tsvector('simple', name) @@ plainto_tsquery('simple', ?)
		  )` + roleClause + `
		ORDER BY (staff_uuid IS NULL), name
		LIMIT ?
	`

	var rows []model.NurseSearchResult
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, args...).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgSearchNurses)
	}
	return rows, nil
}

// MissingDoctorIDs looks up existing active doctor ids in the institution and
// returns those from the input set that weren't found. The diagnosis usecase
// turns the missing list into per-field 422 messages before opening a TX.
func (c *Conn) MissingDoctorIDs(ctx context.Context, institutionID int64, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, institutionID)
	for _, id := range ids {
		args = append(args, id)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	sql := `
		SELECT id FROM mdl_mst_doctor
		WHERE institution_id = ?
		  AND active = TRUE
		  AND id IN (` + placeholders + `)
	`

	var found []string
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, args...).Find(&found)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgMissingDoctorID)
	}

	foundSet := make(map[string]struct{}, len(found))
	for _, id := range found {
		foundSet[id] = struct{}{}
	}
	var missing []string
	for _, id := range ids {
		if _, ok := foundSet[id]; !ok {
			missing = append(missing, id)
		}
	}
	return missing, nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultSearchLimit
	}
	if limit > maxSearchLimit {
		return maxSearchLimit
	}
	return limit
}
