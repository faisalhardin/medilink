-- ICD-9-CM procedure code reference table (Indonesian healthcare standard).
-- Seeded from docs/ICD9 CM.csv via scripts/icd9cm_csv_to_sql.py.
-- Global reference data — not institution-scoped.
CREATE TABLE IF NOT EXISTS mdl_ref_icd9cm (
    code        VARCHAR(10)  PRIMARY KEY,
    display     VARCHAR(500) NOT NULL,
    parent_code VARCHAR(10),
    depth       SMALLINT     NOT NULL,
    is_leaf     BOOLEAN      NOT NULL DEFAULT FALSE,
    version     VARCHAR(20)  NOT NULL DEFAULT 'ICD9CM_2010'
);

CREATE INDEX IF NOT EXISTS idx_mdl_ref_icd9cm_parent_code
    ON mdl_ref_icd9cm (parent_code);

CREATE INDEX IF NOT EXISTS idx_mdl_ref_icd9cm_search
    ON mdl_ref_icd9cm USING gin (to_tsvector('simple', code || ' ' || display));
