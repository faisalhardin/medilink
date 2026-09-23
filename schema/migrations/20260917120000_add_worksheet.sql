-- Worksheet: single-staff wrap for visit commissions.
-- Migrates mdl_trx_visit_commission from period_id → worksheet_id.

DO $$ BEGIN
    CREATE TYPE worksheet_status_enum AS ENUM ('pending', 'open', 'finalized');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE worksheet_generate_status_enum AS ENUM ('idle', 'running', 'succeeded', 'failed');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS mdl_trx_worksheet (
    id                      BIGSERIAL                           PRIMARY KEY,
    uuid                    UUID                                NOT NULL DEFAULT gen_random_uuid(),
    institution_id          BIGINT                              NOT NULL,
    staff_id                UUID                                NOT NULL,
    label                   VARCHAR(100)                        NOT NULL,
    period_start            DATE                                NOT NULL,
    period_end              DATE                                NOT NULL,
    status                  worksheet_status_enum               NOT NULL DEFAULT 'open',
    generate_status         worksheet_generate_status_enum      NOT NULL DEFAULT 'idle',
    compensation_period_id  BIGINT,
    total_commission        BIGINT,
    visit_count             INT,
    generate_started_at     TIMESTAMPTZ,
    generate_finished_at    TIMESTAMPTZ,
    generate_error          TEXT,
    finalized_at            TIMESTAMPTZ,
    finalized_by            UUID,
    created_by              UUID,
    create_time             TIMESTAMPTZ                         NOT NULL DEFAULT NOW(),
    update_time             TIMESTAMPTZ                         NOT NULL DEFAULT NOW(),
    delete_time             TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_worksheet_uuid
    ON mdl_trx_worksheet (uuid);

CREATE INDEX IF NOT EXISTS idx_worksheet_institution_staff_status
    ON mdl_trx_worksheet (institution_id, staff_id, status)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_worksheet_institution_staff_dates
    ON mdl_trx_worksheet (institution_id, staff_id, period_start, period_end)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_worksheet_compensation_period
    ON mdl_trx_worksheet (compensation_period_id)
    WHERE delete_time IS NULL AND compensation_period_id IS NOT NULL;

-- Add worksheet_id nullable for backfill
ALTER TABLE mdl_trx_visit_commission
    ADD COLUMN IF NOT EXISTS worksheet_id BIGINT;

-- One worksheet per (period, staff) that has commission rows
INSERT INTO mdl_trx_worksheet (
    uuid,
    institution_id,
    staff_id,
    label,
    period_start,
    period_end,
    status,
    generate_status,
    compensation_period_id,
    total_commission,
    visit_count,
    create_time,
    update_time
)
SELECT
    gen_random_uuid(),
    p.institution_id,
    c.staff_id,
    p.label,
    p.period_start,
    p.period_end,
    CASE
        WHEN p.status = 'finalized' THEN 'finalized'::worksheet_status_enum
        ELSE 'open'::worksheet_status_enum
    END,
    'succeeded'::worksheet_generate_status_enum,
    p.id,
    COALESCE(SUM(c.commission_amount) FILTER (WHERE c.approved_at IS NOT NULL AND c.delete_time IS NULL), 0),
    COUNT(DISTINCT c.visit_id) FILTER (WHERE c.approved_at IS NOT NULL AND c.delete_time IS NULL)::INT,
    NOW(),
    NOW()
FROM mdl_trx_visit_commission c
JOIN mdl_trx_compensation_period p ON p.id = c.period_id
WHERE c.worksheet_id IS NULL
GROUP BY p.id, p.institution_id, p.label, p.period_start, p.period_end, p.status, c.staff_id;

UPDATE mdl_trx_visit_commission c
SET worksheet_id = w.id
FROM mdl_trx_worksheet w
WHERE c.worksheet_id IS NULL
  AND w.compensation_period_id = c.period_id
  AND w.staff_id = c.staff_id
  AND w.delete_time IS NULL;

-- Orphan commissions (period missing): create placeholder worksheets
INSERT INTO mdl_trx_worksheet (
    uuid, institution_id, staff_id, label, period_start, period_end,
    status, generate_status, create_time, update_time
)
SELECT
    gen_random_uuid(),
    0,
    c.staff_id,
    'migrated-orphan',
    CURRENT_DATE,
    CURRENT_DATE,
    'open'::worksheet_status_enum,
    'succeeded'::worksheet_generate_status_enum,
    NOW(),
    NOW()
FROM mdl_trx_visit_commission c
WHERE c.worksheet_id IS NULL
GROUP BY c.staff_id;

UPDATE mdl_trx_visit_commission c
SET worksheet_id = w.id
FROM mdl_trx_worksheet w
WHERE c.worksheet_id IS NULL
  AND w.staff_id = c.staff_id
  AND w.label = 'migrated-orphan'
  AND w.delete_time IS NULL;

ALTER TABLE mdl_trx_visit_commission
    ALTER COLUMN worksheet_id SET NOT NULL;

DROP INDEX IF EXISTS uidx_commission_period_visit_staff;
DROP INDEX IF EXISTS idx_commission_period_staff;
DROP INDEX IF EXISTS idx_commission_period_unapproved;

ALTER TABLE mdl_trx_visit_commission
    DROP COLUMN IF EXISTS period_id;

CREATE UNIQUE INDEX IF NOT EXISTS uidx_commission_worksheet_visit
    ON mdl_trx_visit_commission (worksheet_id, visit_id)
    WHERE delete_time IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uidx_commission_visit_staff
    ON mdl_trx_visit_commission (visit_id, staff_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_commission_worksheet
    ON mdl_trx_visit_commission (worksheet_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_commission_worksheet_unapproved
    ON mdl_trx_visit_commission (worksheet_id)
    WHERE delete_time IS NULL AND approved_at IS NULL;

ALTER TABLE mdl_trx_patient_visit
    ADD COLUMN IF NOT EXISTS worksheet_id BIGINT NULL;

UPDATE mdl_trx_patient_visit v
SET worksheet_id = sub.worksheet_id
FROM (
    SELECT DISTINCT ON (c.visit_id) c.visit_id, c.worksheet_id
    FROM mdl_trx_visit_commission c
    JOIN mdl_trx_worksheet w ON w.id = c.worksheet_id AND w.status = 'finalized'
    WHERE c.delete_time IS NULL
    ORDER BY c.visit_id, c.worksheet_id
) sub
WHERE v.id = sub.visit_id
  AND v.compensation_locked_at IS NOT NULL
  AND v.worksheet_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_visit_worksheet_lock
    ON mdl_trx_patient_visit (worksheet_id)
    WHERE worksheet_id IS NOT NULL;
