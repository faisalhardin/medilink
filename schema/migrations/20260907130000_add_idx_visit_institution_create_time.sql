-- Period staff-list detection bounds visits by institution + create_time.
CREATE INDEX IF NOT EXISTS idx_trx_patient_visit_institution_create_time
    ON mdl_trx_patient_visit (id_mst_institution, create_time)
    WHERE delete_time IS NULL;
