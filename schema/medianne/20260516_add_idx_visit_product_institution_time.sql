-- Speed up institution-scoped product order statistics queries.
CREATE INDEX IF NOT EXISTS idx_trx_visit_product_institution_create_time
    ON mdl_trx_visit_product (id_mst_institution, create_time)
    WHERE delete_time IS NULL;
