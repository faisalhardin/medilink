CREATE INDEX IF NOT EXISTS idx_mst_patient_institution_name
    ON mdl_mst_patient_institution USING gin (to_tsvector('simple', name));
