CREATE TABLE mdl_mst_product (
  id serial PRIMARY KEY,
  name VARCHAR NOT NULL,
  description VARCHAR,
  added_by VARCHAR NOT NULL,
  create_time TIMESTAMPTZ NOT NULL,
  update_time TIMESTAMPTZ NOT NULL,
  delete_time TIMESTAMPTZ
);

CREATE TABLE mdl_trx_institution_product (
  id serial PRIMARY KEY,
  name VARCHAR NOT NULL,
  id_mst_product BIGINT NULL,
  id_mst_institution BIGINT NOT NULL,
  price numeric NOT NULL default 0,
  is_item BOOLEAN NOT NULL,
  is_treatment BOOLEAN NOT NULL,
  create_time TIMESTAMPTZ NOT NULL,
  update_time TIMESTAMPTZ NOT NULL,
  delete_time TIMESTAMPTZ
);

CREATE TABLE mdl_dtl_institution_product_stock (
  id serial PRIMARY KEY,
  quantity BIGINT NOT NULL,
  unit_type VARCHAR NOT NULL,
  id_mst_institution_product BIGINT NOT NULL,
  create_time TIMESTAMPTZ NOT NULL,
  update_time TIMESTAMPTZ NOT NULL,
  delete_time TIMESTAMPTZ
);

CREATE TABLE mdl_trx_visit_product (
  id serial PRIMARY KEY,
  id_trx_institution_product BIGINT NOT NULL,
  id_trx_patient_visit BIGINT NOT NULL,
  quantity INT NOT NULL,
  unit_type VARCHAR NOT NULL,
  price numeric NOT NULL,
  discount_amount numeric NOT NULL default 0,
  discount_price numeric NOT NULL default 0,
  final_price numeric NOT NULL,
  adjusted_price numeric NULL,
  create_time TIMESTAMPTZ NOT NULL,
  update_time TIMESTAMPTZ NOT NULL,
  delete_time TIMESTAMPTZ
);