-- Compensation domain: staff wages, payday periods, visit commissions,
-- manual visit contributors, and visit lock columns.
--
-- ID strategy (PRD/TRD + live schema):
--   institution_id, visit_id → BIGINT (mdl_mst_institution.id, mdl_trx_patient_visit.id)
--   staff_id / audit staff cols → UUID (mdl_mst_staff.uuid; same as doctor/nurse staff_uuid)
--   mdl_trx_compensation_period.uuid → public API key only
--
-- No physical REFERENCES / ON DELETE (project convention).

-- ---------------------------------------------------------------------------
-- Enums
-- ---------------------------------------------------------------------------
DO $$ BEGIN
    CREATE TYPE wage_cadence_enum AS ENUM ('monthly', 'weekly');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE compensation_period_status_enum AS ENUM ('open', 'draft', 'finalized');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE commission_type_enum AS ENUM ('percent', 'flat');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ---------------------------------------------------------------------------
-- mdl_mst_staff_wage
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mdl_mst_staff_wage (
    id              BIGSERIAL           PRIMARY KEY,
    staff_id        UUID                NOT NULL,
    institution_id  BIGINT              NOT NULL,
    wage_amount     BIGINT              NOT NULL,
    wage_cadence    wage_cadence_enum   NOT NULL,
    is_active       BOOLEAN             NOT NULL DEFAULT true,
    effective_from  DATE                NOT NULL,
    effective_to    DATE,
    created_by      UUID,
    updated_by      UUID,
    create_time     TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    update_time     TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    delete_time     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_staff_wage_staff_institution
    ON mdl_mst_staff_wage (staff_id, institution_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_staff_wage_effective
    ON mdl_mst_staff_wage (staff_id, institution_id, effective_from, effective_to)
    WHERE delete_time IS NULL;

-- ---------------------------------------------------------------------------
-- mdl_trx_compensation_period
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mdl_trx_compensation_period (
    id              BIGSERIAL                           PRIMARY KEY,
    uuid            UUID                                NOT NULL DEFAULT gen_random_uuid(),
    institution_id  BIGINT                              NOT NULL,
    label           VARCHAR(100)                        NOT NULL,
    period_start    DATE                                NOT NULL,
    period_end      DATE                                NOT NULL,
    status          compensation_period_status_enum     NOT NULL DEFAULT 'open',
    wage_snapshot   JSONB,
    total_wage      BIGINT,
    total_commission BIGINT,
    total_payout    BIGINT,
    staff_count     INT,
    visit_count     INT,
    drafted_at      TIMESTAMPTZ,
    drafted_by      UUID,
    finalized_at    TIMESTAMPTZ,
    finalized_by    UUID,
    create_time     TIMESTAMPTZ                         NOT NULL DEFAULT NOW(),
    update_time     TIMESTAMPTZ                         NOT NULL DEFAULT NOW(),
    delete_time     TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_comp_period_uuid
    ON mdl_trx_compensation_period (uuid);

CREATE INDEX IF NOT EXISTS idx_comp_period_institution_status
    ON mdl_trx_compensation_period (institution_id, status)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_comp_period_dates
    ON mdl_trx_compensation_period (institution_id, period_start, period_end)
    WHERE delete_time IS NULL;

-- ---------------------------------------------------------------------------
-- mdl_trx_visit_commission
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mdl_trx_visit_commission (
    id                      BIGSERIAL               PRIMARY KEY,
    period_id               BIGINT                  NOT NULL,
    visit_id                BIGINT                  NOT NULL,
    staff_id                UUID                    NOT NULL,
    revenue_base            BIGINT                  NOT NULL,
    commission_type         commission_type_enum    NOT NULL,
    commission_percent      DECIMAL(5,2),
    commission_flat_amount  BIGINT,
    commission_amount       BIGINT                  NOT NULL,
    sources                 JSONB,
    note                    TEXT,
    included_manually       BOOLEAN                 NOT NULL DEFAULT false,
    create_time             TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    update_time             TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    delete_time             TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_commission_period_visit_staff
    ON mdl_trx_visit_commission (period_id, visit_id, staff_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_commission_period_staff
    ON mdl_trx_visit_commission (period_id, staff_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_commission_visit
    ON mdl_trx_visit_commission (visit_id)
    WHERE delete_time IS NULL;

-- ---------------------------------------------------------------------------
-- mdl_map_visit_contributor
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mdl_map_visit_contributor (
    id              BIGSERIAL       PRIMARY KEY,
    visit_id        BIGINT          NOT NULL,
    staff_id        UUID            NOT NULL,
    institution_id  BIGINT          NOT NULL,
    added_by        UUID            NOT NULL,
    create_time     TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    delete_time     TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_contributor_visit_staff
    ON mdl_map_visit_contributor (visit_id, staff_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_contributor_visit
    ON mdl_map_visit_contributor (visit_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_contributor_staff
    ON mdl_map_visit_contributor (staff_id, institution_id)
    WHERE delete_time IS NULL;

-- ---------------------------------------------------------------------------
-- Alter mdl_trx_patient_visit — compensation lock columns
-- ---------------------------------------------------------------------------
ALTER TABLE mdl_trx_patient_visit
    ADD COLUMN IF NOT EXISTS compensation_period_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS compensation_locked_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_visit_compensation_lock
    ON mdl_trx_patient_visit (compensation_locked_at)
    WHERE compensation_locked_at IS NOT NULL;
