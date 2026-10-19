-- Generating is generate_status, not a worksheet lifecycle status.
UPDATE mdl_trx_worksheet
SET status = 'open'
WHERE status = 'pending'
  AND delete_time IS NULL;
