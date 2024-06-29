-- public.av_mst_user definition

-- Drop table

-- DROP TABLE public.av_mst_user;

CREATE TABLE public.av_mst_user (
	id int4 NOT NULL,
	"uuid" varchar NULL,
	email varchar NULL,
	"domain" varchar NULL,
	create_time timestamptz NULL,
	update_time timestamptz NULL,
	delete_time timestamptz NULL,
	CONSTRAINT av_mst_user_pkey PRIMARY KEY (id)
);

CREATE TABLE "av_mst_role" (
  "id" integer PRIMARY KEY,
  "title" varchar,
  "description" text,
  "create_time" timestamp,
  "update_time" timestamp,
  "delete_time" timestamp
);

CREATE TABLE "av_map_user_role" (
  "id_mst_role" integer,
  "id_mst_user" integer
);
