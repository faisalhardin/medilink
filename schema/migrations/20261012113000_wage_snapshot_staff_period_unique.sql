-- One live wage snapshot per staff and compensation period.
-- Older duplicates are soft-deleted so the unique index can be created.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY staff_id, compensation_period_id
               ORDER BY id DESC
           ) AS rn
    FROM mdl_trx_wage_period_snapshot
    WHERE delete_time IS NULL
)
UPDATE mdl_trx_wage_period_snapshot AS s
SET delete_time = NOW()
FROM ranked
WHERE s.id = ranked.id
  AND ranked.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS uidx_wage_period_snapshot_staff_period
    ON mdl_trx_wage_period_snapshot (staff_id, compensation_period_id)
    WHERE delete_time IS NULL;
