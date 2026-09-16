-- Create odontogram history table for event sourcing with CRDT support
CREATE TABLE IF NOT EXISTS public.mdl_hst_odontogram (
    event_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id int8 not null,
    patient_id int8 NOT NULL,
    visit_id int8 NOT NULL,
    journey_point_short_id varchar,
    event_type varchar NOT NULL,
    tooth_id varchar(2) NOT NULL,
    sequence_number int8 NOT NULL,
    event_data jsonb NOT NULL,
    logical_timestamp int8 NOT NULL,
    created_by_staff_id int8 NOT NULL,
    unix_timestamp int8 NOT NULL,
    created_by varchar NOT NULL,
    create_time int8 NOT NULL,
    
    CONSTRAINT mdl_hst_odontogram_patient_sequence_unique UNIQUE (patient_id, sequence_number)
);

-- Indexes for efficient querying
CREATE INDEX idx_hst_odontogram_patient_sequence ON public.mdl_hst_odontogram(institution_id, patient_id, sequence_number);
CREATE INDEX idx_hst_odontogram_patient_timestamp ON public.mdl_hst_odontogram(institution_id, patient_id, logical_timestamp, created_by_staff_id);
CREATE INDEX idx_hst_odontogram_visit ON public.mdl_hst_odontogram(institution_id, visit_id);
CREATE INDEX idx_hst_odontogram_tooth ON public.mdl_hst_odontogram(institution_id, patient_id, tooth_id, sequence_number);

-- Create odontogram snapshot table for caching built state
CREATE TABLE IF NOT EXISTS public.mdl_mst_patient_odontogram (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id int8 not null,
    patient_id int8 NOT NULL UNIQUE,
    snapshot jsonb NOT NULL,
    last_event_sequence int8 NOT NULL,
    max_logical_timestamp int8 NOT NULL,
    last_updated int8 NOT NULL
);

-- Index for patient lookup
CREATE INDEX idx_mst_patient_odontogram_patient ON public.mdl_mst_patient_odontogram(institution_id, patient_id);

