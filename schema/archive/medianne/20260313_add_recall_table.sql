-- Recall: scheduled control or future appointment reminder for doctors
CREATE TABLE IF NOT EXISTS public.mdl_trx_recall (
    id serial4 NOT NULL,
    id_mst_patient int8 NOT NULL,
    id_mst_institution int8 NOT NULL,
    scheduled_at timestamptz NOT NULL,
    recall_type varchar NOT NULL,
    notes varchar NULL,
    created_by_id_mst_staff int8 NOT NULL,
    id_trx_patient_visit int8 NULL,
    create_time timestamptz NOT NULL,
    update_time timestamptz NOT NULL,
    delete_time timestamptz NULL,
    CONSTRAINT mdl_trx_recall_pkey PRIMARY KEY (id)
);

CREATE INDEX idx_trx_recall_institution_scheduled ON public.mdl_trx_recall(id_mst_institution, scheduled_at) WHERE delete_time IS NULL;
CREATE INDEX idx_trx_recall_patient ON public.mdl_trx_recall(id_mst_patient, scheduled_at) WHERE delete_time IS NULL;
