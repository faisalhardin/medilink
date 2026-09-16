-- Membership rows from generate are unapproved until PUT sets approved_at.
ALTER TABLE mdl_trx_visit_commission
    ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_commission_period_unapproved
    ON mdl_trx_visit_commission (period_id)
    WHERE delete_time IS NULL AND approved_at IS NULL;
