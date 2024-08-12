CREATE TABLE mdl_mst_institution (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    staff_number INT NOT NULL,
    max_staff INT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL,
    update_time TIMESTAMPTZ NOT NULL,
    delete_time TIMESTAMPTZ
);