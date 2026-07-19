package procedure

import (
	"context"
	"strings"
	"unicode"

	"github.com/faisalhardin/medilink/internal/entity/model"
	procedurerepo "github.com/faisalhardin/medilink/internal/entity/repo/procedure"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
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

// sanitizeTsQuery converts a raw user input string into a safe PostgreSQL
// to_tsquery expression with prefix matching on every token.
//
// Steps:
//  1. Strip characters that have special meaning in tsquery syntax.
//  2. Collapse whitespace and split into tokens.
//  3. Drop tokens shorter than 2 characters (noise like lone "s").
//  4. Join tokens with " & " (AND) and append ":*" for prefix matching.
//
// Examples:
//
//	"append"        → "append:*"
//	"knee arthro"   → "knee:* & arthro:*"
//	"lapar(oscopy)" → "lapar:* & oscopy:*"
//	"doctor's"      → "doctor:*"          (lone "s" is filtered out)
//	""              → ""                  (caller should guard against empty)
func sanitizeTsQuery(input string) string {
	// Replace characters special to tsquery (&|!()<>:'\-) with a space.
	// The hyphen is also replaced because after stripping angle brackets from
	// phrase-operator syntax like "<->" only a bare "-" remains, which is not
	// a valid tsquery lexeme.
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		switch r {
		case '&', '|', '!', '(', ')', '<', '>', ':', '\'', '\\', '-':
			b.WriteRune(' ')
		default:
			if unicode.IsPrint(r) {
				b.WriteRune(r)
			}
		}
	}

	// Split on whitespace, drop empty or single-character noise tokens.
	parts := strings.Fields(b.String())
	terms := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) >= 2 {
			terms = append(terms, p+":*")
		}
	}

	if len(terms) == 0 {
		return ""
	}
	return strings.Join(terms, " & ")
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

	tsq := sanitizeTsQuery(q)

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
	//   tier 3 – FTS match, weighted by ts_rank_cd (lower = better)
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
		        ELSE 1
		    END ASC,
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
