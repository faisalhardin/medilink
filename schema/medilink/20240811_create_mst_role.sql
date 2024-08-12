CREATE TABLE mdl_mst_role (
  id SERIAL PRIMARY KEY,
  role_id INTEGER,
  name VARCHAR,
  create_time TIMESTAMPTZ,
  update_time TIMESTAMPTZ,
  delete_time TIMESTAMPTZ
);