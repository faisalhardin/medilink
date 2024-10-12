CREATE TABLE mdl_dtl_patient_visit (
  id serial PRIMARY KEY,
  id_trx_patient_visit bigint not null,
  touchpoint_name varchar,
  touchpoint_category varchar,
  notes varchar,
  create_time TIMESTAMPTZ not null,
  update_time TIMESTAMPTZ not null,
  delete_time TIMESTAMPTZ
);

alter table mdl_trx_patient_visit add column id_mst_institution bigint not null default 0;

with mmpi as (select id, id_mst_institution from mdl_mst_patient_institution)
update  mdl_trx_patient_visit
set id_mst_institution = mmpi.id_mst_institution
from mmpi
where id_mst_patient = mmpi.id;