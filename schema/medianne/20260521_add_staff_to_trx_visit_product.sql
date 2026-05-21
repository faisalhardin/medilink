-- Staff accountability on visit product lines.
ALTER TABLE public.mdl_trx_visit_product
    ADD COLUMN IF NOT EXISTS id_mst_staff_created_by int8 NULL,
    ADD COLUMN IF NOT EXISTS id_mst_staff_updated_by int8 NULL;
