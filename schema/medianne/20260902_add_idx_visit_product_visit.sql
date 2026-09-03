-- Speed up full-visit revenue_base aggregation for compensation (Option A).
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trx_visit_product_visit
    ON mdl_trx_visit_product (id_trx_patient_visit)
    WHERE delete_time IS NULL;
