-- Existing rows keep their generate time. Later edits write update_time.
ALTER TABLE mdl_trx_wage_period_snapshot
    ADD COLUMN IF NOT EXISTS update_time TIMESTAMPTZ;

UPDATE mdl_trx_wage_period_snapshot
SET update_time = create_time
WHERE update_time IS NULL;

ALTER TABLE mdl_trx_wage_period_snapshot
    ALTER COLUMN update_time SET DEFAULT NOW(),
    ALTER COLUMN update_time SET NOT NULL;
