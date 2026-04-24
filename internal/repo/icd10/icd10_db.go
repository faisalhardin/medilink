package icd10

import (
	"context"
	"strings"

	"github.com/faisalhardin/medilink/internal/entity/model"
	icd10repo "github.com/faisalhardin/medilink/internal/entity/repo/icd10"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix    = "ICD10DB."
	WrapMsgSearch       = WrapErrMsgPrefix + "Search"
	WrapMsgMissingCodes = WrapErrMsgPrefix + "MissingCodes"

	defaultSearchLimit = 10
	maxSearchLimit     = 50
)

type Conn struct {
	DB *xormlib.DBConnect
}

// NewICD10DB returns an ICD10DB implementation bound to the given xorm connection.
func NewICD10DB(db *xormlib.DBConnect) icd10repo.ICD10DB {
	return &Conn{DB: db}
}

// Search runs an ILIKE-prefix match on `code` unioned with a tsvector
// match on `display`; prefix hits are ordered first so autocomplete feels right.
func (c *Conn) Search(ctx context.Context, q string, limit int) ([]model.RefICD10, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []model.RefICD10{}, nil
	}
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	const sql = `
		SELECT code, display, category, created_at
		FROM ref_icd10
		WHERE code ILIKE ? || '%'
		   OR to_tsvector('simple', display) @@ plainto_tsquery('simple', ?)
		ORDER BY CASE WHEN code ILIKE ? || '%' THEN 0 ELSE 1 END, code
		LIMIT ?
	`

	var rows []model.RefICD10
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, q, q, q, limit).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgSearch)
	}
	return rows, nil
}

// MissingCodes fetches only the codes that EXIST, then diffs in memory to
// return those that don't. Avoids per-code round trips and keeps the DB
// query simple enough for xorm without raw array binding.
func (c *Conn) MissingCodes(ctx context.Context, codes []string) ([]string, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	args := make([]interface{}, 0, len(codes))
	for _, code := range codes {
		args = append(args, code)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(codes)), ",")
	sql := "SELECT code FROM ref_icd10 WHERE code IN (" + placeholders + ")"

	var found []string
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, args...).Find(&found)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgMissingCodes)
	}

	foundSet := make(map[string]struct{}, len(found))
	for _, code := range found {
		foundSet[code] = struct{}{}
	}

	var missing []string
	for _, code := range codes {
		if _, ok := foundSet[code]; !ok {
			missing = append(missing, code)
		}
	}
	return missing, nil
}
