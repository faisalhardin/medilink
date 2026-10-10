-- Wage snapshots generated for a payday period from active wage contracts.
-- period_start and period_end are copies of that period's dates.
-- Nullable prorate inputs and stored wage totals: NULL means the value was
-- not entered. Zero is a real amount or day count.

CREATE TABLE IF NOT EXISTS mdl_trx_wage_period_snapshot (
    id                       BIGSERIAL           PRIMARY KEY,
    staff_id                 UUID                NOT NULL,
    institution_id           BIGINT              NOT NULL,
    compensation_period_id   BIGINT              NOT NULL,
    period_start             DATE                NOT NULL,
    period_end               DATE                NOT NULL,
    wage_amount              BIGINT              NOT NULL,
    wage_cadence             wage_cadence_enum   NOT NULL,
    mandatory_working_days   INT,
    staff_working_days       INT,
    final_wage               BIGINT,
    total_wage               BIGINT,
    create_time              TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    delete_time              TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_wage_period_snapshot_institution
    ON mdl_trx_wage_period_snapshot (institution_id, id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_wage_period_snapshot_period
    ON mdl_trx_wage_period_snapshot (institution_id, compensation_period_id)
    WHERE delete_time IS NULL;
