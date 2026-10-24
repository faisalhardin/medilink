package procedure

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	procedurerepo "github.com/faisalhardin/medilink/internal/entity/repo/procedure"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/util/fts"
	"github.com/pkg/errors"
)

const (
	WrapMsgICD9CMSearch    = "ICD9CMDB.Search"
	WrapMsgICD9CMGetByCode = "ICD9CMDB.GetByCode"

	defaultICD9CMLimit = 10
	maxICD9CMLimit     = 50
)

type ICD9CMConn struct {
	DB *xormlib.DBConnect
}

// NewICD9CMDB returns an ICD9CMDB bound to the given xorm connection.
func NewICD9CMDB(db *xormlib.DBConnect) procedurerepo.ICD9CMDB {
	return &ICD9CMConn{DB: db}
}

// Search returns up to limit rows matching q by:
//   - exact code match (ranked 1st)
//   - code prefix match (ranked 2nd)
//   - display full-text prefix match, scored by ts_rank_cd (ranked 3rd+)
//
// Leaf nodes are promoted within each rank tier to surface billable codes.
func (c *ICD9CMConn) Search(ctx context.Context, q string, limit int) ([]model.ICD9CMOption, error) {
	if limit <= 0 {
		limit = defaultICD9CMLimit
	}
	if limit > maxICD9CMLimit {
		limit = maxICD9CMLimit
	}

	tsq := fts.SanitizeTsQuery(q)

	// When sanitization strips everything (e.g. input was only special chars),
	// fall back to a code-prefix-only search so callers always get a result set.
	if tsq == "" {
		const fallbackSQL = `
			SELECT code, display, depth, is_leaf
			FROM mdl_ref_icd9cm
			WHERE code ILIKE ? || '%'
			ORDER BY is_leaf DESC, code ASC
			LIMIT ?
		`
		var rows []model.ICD9CMOption
		err := c.DB.SlaveDB.Context(ctx).SQL(fallbackSQL, q, limit).Find(&rows)
		if err != nil {
			return nil, errors.Wrap(err, WrapMsgICD9CMSearch)
		}
		return rows, nil
	}

	// Main query uses a lateral sub-expression to materialise the tsquery once,
	// then ranks results in three tiers:
	//   tier 1 – exact code equality
	//   tier 2 – code ILIKE prefix
	//   tier 3 – FTS match, weighted by ts_rank_cd (higher = better match)
	//
	// The @@ predicate uses to_tsvector('simple', code || ' ' || display) which
	// exactly matches the expression in idx_mdl_ref_icd9cm_search, so PostgreSQL
	// can use the GIN index instead of doing a sequential scan.
	const searchSQL = `
		SELECT code, display, depth, is_leaf
		FROM mdl_ref_icd9cm,
		     to_tsquery('simple', ?) AS q
		WHERE code ILIKE ? || '%'
		   OR to_tsvector('simple', code || ' ' || display) @@ q
		ORDER BY
		    CASE
		        WHEN code = ?            THEN 0
		        WHEN code ILIKE ? || '%' THEN 1
		        ELSE 2
		    END ASC,
		    CASE
		        WHEN to_tsvector('simple', code || ' ' || display) @@ q
		        THEN ts_rank_cd(to_tsvector('simple', code || ' ' || display), q)
		        ELSE 0
		    END DESC,
		    is_leaf DESC,
		    code ASC
		LIMIT ?
	`

	var rows []model.ICD9CMOption
	// Args: tsq, q (ILIKE), q (exact), q (ILIKE in ORDER BY), limit
	err := c.DB.SlaveDB.Context(ctx).SQL(searchSQL, tsq, q, q, q, limit).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgICD9CMSearch)
	}
	return rows, nil
}

// GetByCode returns the single row with the exact code, or nil if not found.
func (c *ICD9CMConn) GetByCode(ctx context.Context, code string) (*model.RefICD9CM, error) {
	const sql = `
		SELECT code, display, parent_code, depth, is_leaf, version
		FROM mdl_ref_icd9cm
		WHERE code = ?
	`

	var row model.RefICD9CM
	found, err := c.DB.SlaveDB.Context(ctx).SQL(sql, code).Get(&row)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgICD9CMGetByCode)
	}
	if !found {
		return nil, nil
	}
	return &row, nil
}
