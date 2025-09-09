-- public.mdl_dtl_institution_product_stock definition

-- Drop table

-- DROP TABLE public.mdl_dtl_institution_product_stock;

CREATE TABLE public.mdl_dtl_institution_product_stock (
	id serial4 NOT NULL,
	quantity int8 NOT NULL,
	unit_type varchar NOT NULL,
	id_trx_institution_product int8 NOT NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	CONSTRAINT mdl_dtl_institution_product_stock_pkey PRIMARY KEY (id)
);


-- public.mdl_dtl_patient_visit definition

-- Drop table

-- DROP TABLE public.mdl_dtl_patient_visit;

CREATE TABLE public.mdl_dtl_patient_visit (
	id serial4 NOT NULL,
	id_trx_patient_visit int8 NOT NULL,
	name_mst_journey_point varchar NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	id_mst_journey_point int8 DEFAULT 0 NOT NULL,
	action_by_id_mst_staff int8 NOT NULL,
	notes jsonb NULL,
	contributors jsonb NULL,
	id_mst_service_point int8 NULL,
	CONSTRAINT mdl_dtl_patient_visit_pkey PRIMARY KEY (id)
);


-- public.mdl_map_role_staff definition

-- Drop table

-- DROP TABLE public.mdl_map_role_staff;

CREATE TABLE public.mdl_map_role_staff (
	id serial4 NOT NULL,
	id_mst_staff int8 NOT NULL,
	id_mst_role int8 NOT NULL,
	CONSTRAINT mdl_map_role_staff_pkey PRIMARY KEY (id)
);


-- public.mdl_map_service_point_journey_point definition

-- Drop table

-- DROP TABLE public.mdl_map_service_point_journey_point;

CREATE TABLE public.mdl_map_service_point_journey_point (
	id_mst_service_point int8 NULL,
	id_mst_journey_point int8 NULL
);


-- public.mdl_map_staff_journey_point definition

-- Drop table

-- DROP TABLE public.mdl_map_staff_journey_point;

CREATE TABLE public.mdl_map_staff_journey_point (
	id_mst_staff int4 NULL,
	id_mst_journey_point int4 NULL
);


-- public.mdl_map_staff_service_point definition

-- Drop table

-- DROP TABLE public.mdl_map_staff_service_point;

CREATE TABLE public.mdl_map_staff_service_point (
	id_mst_staff int4 NULL,
	id_mst_service_point int4 NULL
);


-- public.mdl_mst_institution definition

-- Drop table

-- DROP TABLE public.mdl_mst_institution;

CREATE TABLE public.mdl_mst_institution (
	id serial4 NOT NULL,
	"name" varchar NOT NULL,
	staff_number int4 NOT NULL,
	max_staff int4 NOT NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	CONSTRAINT mdl_mst_institution_pkey PRIMARY KEY (id)
);


-- public.mdl_mst_journey_board definition

-- Drop table

-- DROP TABLE public.mdl_mst_journey_board;

CREATE TABLE public.mdl_mst_journey_board (
	id serial4 NOT NULL,
	"name" varchar NOT NULL,
	id_mst_institution int8 NOT NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	CONSTRAINT mdl_mst_journey_board_pkey PRIMARY KEY (id)
);


-- public.mdl_mst_journey_point definition

-- Drop table

-- DROP TABLE public.mdl_mst_journey_point;

CREATE TABLE public.mdl_mst_journey_point (
	id serial4 NOT NULL,
	"name" varchar NULL,
	"position" int4 NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	id_mst_journey_board int8 DEFAULT 0 NOT NULL,
	id_mst_institution int8 DEFAULT 0 NOT NULL,
	CONSTRAINT mdl_mst_journey_point_pkey PRIMARY KEY (id)
);


-- public.mdl_mst_patient_institution definition

-- Drop table

-- DROP TABLE public.mdl_mst_patient_institution;

CREATE TABLE public.mdl_mst_patient_institution (
	id serial4 NOT NULL,
	"uuid" varchar DEFAULT uuid_generate_v4() NOT NULL,
	nik varchar NULL,
	"name" varchar NOT NULL,
	place_of_birth varchar NULL,
	date_of_birth timestamp NOT NULL,
	address varchar NULL,
	id_mst_institution int4 NOT NULL,
	religion varchar NULL,
	create_time timestamptz NULL,
	update_time timestamptz NULL,
	delete_time timestamptz NULL,
	sex varchar DEFAULT 'male'::character varying NOT NULL,
	phone_number varchar NULL,
	CONSTRAINT mdl_mst_patient_institution_pkey PRIMARY KEY (id)
);


-- public.mdl_mst_product definition

-- Drop table

-- DROP TABLE public.mdl_mst_product;

CREATE TABLE public.mdl_mst_product (
	id serial4 NOT NULL,
	"name" varchar NOT NULL,
	description varchar NULL,
	added_by varchar NOT NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	CONSTRAINT mdl_mst_product_pkey PRIMARY KEY (id)
);


-- public.mdl_mst_role definition

-- Drop table

-- DROP TABLE public.mdl_mst_role;

CREATE TABLE public.mdl_mst_role (
	id serial4 NOT NULL,
	role_id int4 NULL,
	"name" varchar NULL,
	create_time timestamptz NULL,
	update_time timestamptz NULL,
	delete_time timestamptz NULL,
	CONSTRAINT mdl_mst_role_pkey PRIMARY KEY (id)
);

INSERT INTO public.mdl_mst_role
(role_id, "name", create_time, update_time)
VALUES(1, 'administrator', now(), now());

INSERT INTO public.mdl_mst_role
(role_id, "name", create_time, update_time)
VALUES(2, 'clerk', now(), now());

INSERT INTO public.mdl_mst_role
(role_id, "name", create_time, update_time)
VALUES(3, 'doctor', now(), now());

INSERT INTO public.mdl_mst_role
(role_id, "name", create_time, update_time)
VALUES(4, 'nurse', now(), now());


-- public.mdl_mst_service_point definition

-- Drop table

-- DROP TABLE public.mdl_mst_service_point;

CREATE TABLE public.mdl_mst_service_point (
	id serial4 NOT NULL,
	"name" varchar NOT NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	id_mst_journey_board int8 DEFAULT 0 NOT NULL,
	id_mst_institution int8 DEFAULT 0 NOT NULL,
	CONSTRAINT mdl_mst_service_point_pkey PRIMARY KEY (id)
);


-- public.mdl_mst_staff definition

-- Drop table

-- DROP TABLE public.mdl_mst_staff;

CREATE TABLE public.mdl_mst_staff (
	id serial4 NOT NULL,
	"uuid" varchar DEFAULT uuid_generate_v4() NOT NULL,
	"name" varchar NOT NULL,
	email varchar NOT NULL,
	id_mst_institution int4 NOT NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	CONSTRAINT mdl_mst_staff_pkey PRIMARY KEY (id)
);


-- public.mdl_trx_institution_product definition

-- Drop table

-- DROP TABLE public.mdl_trx_institution_product;

CREATE TABLE public.mdl_trx_institution_product (
	id serial4 NOT NULL,
	"name" varchar NOT NULL,
	id_mst_product int8 NULL,
	id_mst_institution int8 NOT NULL,
	price numeric DEFAULT 0 NOT NULL,
	is_item bool NOT NULL,
	is_treatment bool NOT NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	CONSTRAINT mdl_trx_institution_product_pkey PRIMARY KEY (id)
);


-- public.mdl_trx_patient_visit definition

-- Drop table

-- DROP TABLE public.mdl_trx_patient_visit;

CREATE TABLE public.mdl_trx_patient_visit (
	id serial4 NOT NULL,
	id_mst_patient int8 NOT NULL,
	"action" varchar NULL,
	status varchar NULL,
	notes varchar NULL,
	create_time timestamptz NULL,
	update_time timestamptz NULL,
	delete_time timestamptz NULL,
	id_mst_institution int8 DEFAULT 0 NOT NULL,
	id_mst_journey_board int8 NOT NULL,
	id_mst_journey_point int8 NULL,
	id_mst_service_point int8 NULL,
	mst_journey_point_id_update_unix_time int8 DEFAULT EXTRACT(epoch FROM now()) NOT NULL,
	product_cart jsonb DEFAULT '[]'::jsonb NULL,
	CONSTRAINT mdl_trx_patient_visit_pkey PRIMARY KEY (id)
);


-- public.mdl_trx_visit_product definition

-- Drop table

-- DROP TABLE public.mdl_trx_visit_product;

CREATE TABLE public.mdl_trx_visit_product (
	id serial4 NOT NULL,
	id_trx_institution_product int8 NOT NULL,
	id_trx_patient_visit int8 NOT NULL,
	quantity int4 NOT NULL,
	unit_type varchar NOT NULL,
	price numeric NOT NULL,
	discount_rate numeric DEFAULT 0 NOT NULL,
	discount_price numeric DEFAULT 0 NOT NULL,
	total_price numeric NOT NULL,
	adjusted_price numeric NULL,
	create_time timestamptz NOT NULL,
	update_time timestamptz NOT NULL,
	delete_time timestamptz NULL,
	id_mst_institution int8 DEFAULT 0 NOT NULL,
	id_dtl_patient_visit int8 DEFAULT 0 NOT NULL,
	"name" varchar NULL,
	CONSTRAINT mdl_trx_visit_product_pkey PRIMARY KEY (id)
);