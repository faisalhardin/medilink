-- Visit procedure transaction: one row per procedure recorded in a visit.
-- product_id optionally references mdl_trx_institution_product (is_treatment = true).
-- Snapshot columns (product_name, doctor_name, nurse_name, icd10pcs_display)
-- are captured at write time so reads never need JOINs and history is preserved
-- even if master data changes.
CREATE TABLE IF NOT EXISTS mdl_trx_visit_procedure (
    id                  BIGSERIAL       PRIMARY KEY,
    visit_id            BIGINT          NOT NULL,
    institution_id      BIGINT          NOT NULL,
    product_id          BIGINT,
    product_name        VARCHAR(255),
    doctor_id           VARCHAR(50)     NOT NULL,
    doctor_name         VARCHAR(255)    NOT NULL,
    nurse_id            VARCHAR(50),
    nurse_name          VARCHAR(255),
    planned_at          TIMESTAMPTZ,
    category            VARCHAR(20),
    duration            VARCHAR(100),
    icd9cm_code         VARCHAR(10),
    icd9cm_display      VARCHAR(500),
    description         TEXT,
    notes               TEXT,
    rank                SMALLINT        NOT NULL CHECK (rank >= 1),
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_trx_visit_procedure_visit_active
    ON mdl_trx_visit_procedure (institution_id, visit_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_trx_visit_procedure_inst_id_active
    ON mdl_trx_visit_procedure (institution_id, id)
    WHERE deleted_at IS NULL;
