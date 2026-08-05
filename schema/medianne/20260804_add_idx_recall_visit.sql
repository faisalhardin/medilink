CREATE INDEX IF NOT EXISTS idx_trx_recall_visit
    ON public.mdl_trx_recall(id_trx_patient_visit)
    WHERE delete_time IS NULL AND id_trx_patient_visit IS NOT NULL;
