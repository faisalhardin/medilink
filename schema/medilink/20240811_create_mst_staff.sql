

CREATE TABLE mdl_mst_staff (
    id serial PRIMARY KEY,
	uuid varchar not null default uuid_generate_v4(),
    name VARCHAR NOT NULL,
    email VARCHAR not null,
    id_mst_institution INT NOT NULL,
    create_time TIMESTAMPTZ NOT NULL,
    update_time TIMESTAMPTZ NOT NULL,
    delete_time TIMESTAMPTZ
);