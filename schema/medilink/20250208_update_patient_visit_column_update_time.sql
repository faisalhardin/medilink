alter table mdl_trx_patient_visit add column  mst_journey_point_id_update_unix_time bigint not null default EXTRACT(EPOCH FROM NOW());