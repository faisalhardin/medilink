UPDATE mdl_trx_worksheet w
SET delete_time = NOW(),
    update_time = NOW()
WHERE w.delete_time IS NULL
  AND (
    w.compensation_period_id IS NULL
    OR EXISTS (
      SELECT 1
      FROM mdl_trx_compensation_period p
      WHERE p.id = w.compensation_period_id
        AND p.delete_time IS NOT NULL
    )
  );
