alter table mdl_trx_visit_product  add column id_mst_institution bigint not null default 0;

ALTER TABLE mdl_trx_visit_product
RENAME COLUMN discount_amount TO discount_rate;

ALTER TABLE mdl_trx_visit_product
RENAME COLUMN final_price TO total_price;