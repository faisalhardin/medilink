package icd10

import (
	"context"
	"strings"

	"github.com/faisalhardin/medilink/internal/entity/model"
	icd10repo "github.com/faisalhardin/medilink/internal/entity/repo/icd10"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/util/fts"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix  = "ICD10DB."
	WrapMsgSearch     = WrapErrMsgPrefix + "Search"
	WrapMsgGetByCodes = WrapErrMsgPrefix + "GetByCodes"

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

// Search returns up to limit rows matching q by:
//   - exact code match (ranked 1st)
//   - code prefix match (ranked 2nd)
//   - display/synonyms full-text prefix match via ts_rank_cd (ranked 3rd+)
//
// The search uses prefix matching (":*") so partial words like "kari" or
// "pulpi" match "karies" or "pulpitis". The WHERE clause matches the GIN
// index expression using icd10_synonyms_to_text() for efficiency.
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

	tsq := fts.SanitizeTsQuery(q)

	if tsq == "" {
		const fallbackSQL = `
			SELECT code, display, display_id, synonyms_id, category, created_at
			FROM ref_icd10
			WHERE code ILIKE $1 || '%'
			ORDER BY code ASC
			LIMIT $2
		`
		var rows []model.RefICD10
		err := c.DB.SlaveDB.Context(ctx).SQL(fallbackSQL, q, limit).Find(&rows)
		if err != nil {
			return nil, errors.Wrap(err, WrapMsgSearch)
		}
		return rows, nil
	}

	// Main query uses lateral sub-expression to materialize tsquery once.
	// Ranks results in tiers:
	//   tier 0 – exact code equality (case-insensitive)
	//   tier 1 – code ILIKE prefix
	//   tier 2 – FTS match, weighted by ts_rank_cd (higher = better)
	//
	// The @@ predicate uses the same expression as idx_ref_icd10_search
	// (with icd10_synonyms_to_text wrapper) so PostgreSQL uses the GIN index.
	const searchSQL = `
		SELECT code, display, display_id, synonyms_id, category, created_at
		FROM ref_icd10,
		     to_tsquery('simple', $1) AS q
		WHERE code ILIKE $2 || '%'
		   OR to_tsvector('simple', code || ' ' || display || ' ' || COALESCE(display_id, '') || ' ' || COALESCE(icd10_synonyms_to_text(synonyms_id), '')) @@ q
		ORDER BY
		    CASE
		        WHEN UPPER(code) = UPPER($3) THEN 0
		        WHEN code ILIKE $4 || '%' THEN 1
		        ELSE 2
		    END ASC,
		    CASE
		        WHEN to_tsvector('simple', code || ' ' || display || ' ' || COALESCE(display_id, '') || ' ' || COALESCE(icd10_synonyms_to_text(synonyms_id), '')) @@ q
		        THEN ts_rank_cd(to_tsvector('simple', code || ' ' || display || ' ' || COALESCE(display_id, '') || ' ' || COALESCE(icd10_synonyms_to_text(synonyms_id), '')), q)
		        ELSE 0
		    END DESC,
		    code ASC
		LIMIT $5
	`

	var rows []model.RefICD10
	err := c.DB.SlaveDB.Context(ctx).SQL(searchSQL, tsq, q, q, q, limit).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgSearch)
	}
	return rows, nil
}

// GetByCodes returns the ref_icd10 rows for the requested codes. Callers
// build the missing-set in memory by diffing against the request and use
// the display text to snapshot it onto write-side rows.
func (c *Conn) GetByCodes(ctx context.Context, codes []string) ([]model.RefICD10, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	args := make([]interface{}, 0, len(codes))
	for _, code := range codes {
		args = append(args, code)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(codes)), ",")
	sql := "SELECT code, display, category, created_at FROM ref_icd10 WHERE code IN (" + placeholders + ")"

	var rows []model.RefICD10
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, args...).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgGetByCodes)
	}
	return rows, nil
}
