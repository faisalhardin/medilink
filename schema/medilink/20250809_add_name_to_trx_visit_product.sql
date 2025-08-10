ALTER table mdl_trx_visit_product add column name varchar;

ALTER TABLE public.mdl_trx_visit_product
ADD CONSTRAINT uq_mdl_product_visit UNIQUE (id_trx_institution_product, id_trx_patient_visit);