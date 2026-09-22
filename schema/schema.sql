-- Desired-state schema for Atlas migrate diff.
-- Assembled from historical DDL + pending followups (IF NOT EXISTS style).
-- Prefer regenerating with: ./scripts/atlas_inspect_schema.sh (see schema/README.md)


-- ---- 20250908_enable_uuid_extension.sql ----

-- Enable UUID extension for uuid_generate_v4() function
-- This extension is required for tables that use uuid_generate_v4() as default values

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

COMMENT ON EXTENSION "uuid-ossp" IS 'Provides functions to generate UUIDs (universally unique identifiers)';



-- ---- 20250909_init_migration.sql ----

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

-- ---- 20250911_add_short_id_to_journey_point.sql ----

-- Add short_id column to mdl_mst_journey_point table
-- This migration adds a unique short ID field for journey points

ALTER TABLE mdl_mst_journey_point 
ADD COLUMN short_id VARCHAR(8) UNIQUE;

-- Create index for better performance on short_id lookups
CREATE INDEX idx_mdl_mst_journey_point_short_id ON mdl_mst_journey_point(short_id);

-- Add comment to the column
COMMENT ON COLUMN mdl_mst_journey_point.short_id IS 'Short unique identifier for journey point (Base58 encoded, 8 characters)';

ALTER TABLE mdl_mst_journey_point 
ALTER COLUMN short_id SET NOT NULL;

-- ---- 20250922_add_user_sessions.sql ----

-- Create user sessions table for revamped authentication system
CREATE TABLE public.mdl_mst_user_sessions (
    id serial4 NOT NULL,
    session_key varchar NOT NULL UNIQUE,
    user_id int8 NOT NULL,
    access_token_hash varchar NOT NULL,
    refresh_token_hash varchar NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'active',
    expires_at timestamptz NOT NULL,
    refresh_expires_at timestamptz NOT NULL,
    last_accessed_at timestamptz NOT NULL,
    ip_address varchar(45),
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,
    CONSTRAINT mdl_mst_user_sessions_pkey PRIMARY KEY (id),
    CONSTRAINT mdl_mst_user_sessions_status_check CHECK (status IN ('active', 'revoked', 'expired'))
);

-- Create indexes for performance
CREATE INDEX idx_user_sessions_session_key ON mdl_mst_user_sessions(session_key);
CREATE INDEX idx_user_sessions_user_id ON mdl_mst_user_sessions(user_id);
CREATE INDEX idx_user_sessions_status ON mdl_mst_user_sessions(status);
CREATE INDEX idx_user_sessions_expires_at ON mdl_mst_user_sessions(expires_at);
CREATE INDEX idx_user_sessions_refresh_expires_at ON mdl_mst_user_sessions(refresh_expires_at);
CREATE INDEX idx_user_sessions_last_accessed_at ON mdl_mst_user_sessions(last_accessed_at);

-- Add comments to table and columns
COMMENT ON TABLE mdl_mst_user_sessions IS 'User session management for authentication system';
COMMENT ON COLUMN mdl_mst_user_sessions.session_key IS 'Unique session identifier';
COMMENT ON COLUMN mdl_mst_user_sessions.user_id IS 'Reference to staff user';
COMMENT ON COLUMN mdl_mst_user_sessions.access_token_hash IS 'Hashed access token for security';
COMMENT ON COLUMN mdl_mst_user_sessions.refresh_token_hash IS 'Hashed refresh token for security';
COMMENT ON COLUMN mdl_mst_user_sessions.status IS 'Session status: active, revoked, expired';
COMMENT ON COLUMN mdl_mst_user_sessions.expires_at IS 'Access token expiration time';
COMMENT ON COLUMN mdl_mst_user_sessions.refresh_expires_at IS 'Refresh token expiration time';
COMMENT ON COLUMN mdl_mst_user_sessions.last_accessed_at IS 'Last time session was accessed';
COMMENT ON COLUMN mdl_mst_user_sessions.ip_address IS 'Client IP address for security tracking';
COMMENT ON COLUMN mdl_mst_user_sessions.user_agent IS 'Client user agent for security tracking';


-- ---- 20251215_add_odontogram_tables.sql ----

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



-- ---- 20260125_add_patient_occupation.sql ----

ALTER TABLE mdl_mst_patient_institution 
ADD COLUMN occupation VARCHAR NULL;

-- ---- 20260313_add_recall_table.sql ----

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


-- ---- 20260422_add_ref_icd10.sql ----

-- ICD-10 reference table seeded from WHO release / Kemenkes-localised version
CREATE TABLE IF NOT EXISTS ref_icd10 (
    code        VARCHAR(10)  PRIMARY KEY,
    display     VARCHAR(500) NOT NULL,
    category    VARCHAR(100),
    created_at  TIMESTAMP    DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ref_icd10_search
    ON ref_icd10 USING gin(to_tsvector('simple', code || ' ' || display));


-- ---- 20260423_add_mst_doctor.sql ----

-- Doctor master — clinical identity, decoupled from login account
CREATE TABLE IF NOT EXISTS mdl_mst_doctor (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_uuid      UUID         UNIQUE,
    name            VARCHAR(200) NOT NULL,
    sip_number      VARCHAR(50),
    specialization  VARCHAR(100),
    institution_id  BIGINT       NOT NULL,
    active          BOOLEAN      DEFAULT TRUE,
    created_at      TIMESTAMP    DEFAULT NOW(),
    updated_at      TIMESTAMP    DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mst_doctor_name
    ON mdl_mst_doctor USING gin(to_tsvector('simple', name));

CREATE INDEX IF NOT EXISTS idx_mst_doctor_staff_uuid
    ON mdl_mst_doctor(staff_uuid) WHERE staff_uuid IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_mst_doctor_institution
    ON mdl_mst_doctor(institution_id, active);


-- ---- 20260424_add_mst_nurse.sql ----

-- Nurse / midwife / paramedic master
DO $$ BEGIN
    CREATE TYPE nurse_role_type AS ENUM ('nurse', 'midwife', 'paramedic');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS mdl_mst_nurse (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_uuid      UUID            UNIQUE,
    name            VARCHAR(200)    NOT NULL,
    sip_number      VARCHAR(50),
    role            nurse_role_type NOT NULL DEFAULT 'nurse',
    institution_id  BIGINT          NOT NULL,
    active          BOOLEAN         DEFAULT TRUE,
    created_at      TIMESTAMP       DEFAULT NOW(),
    updated_at      TIMESTAMP       DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mst_nurse_name
    ON mdl_mst_nurse USING gin(to_tsvector('simple', name));

CREATE INDEX IF NOT EXISTS idx_mst_nurse_staff_uuid
    ON mdl_mst_nurse(staff_uuid) WHERE staff_uuid IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_mst_nurse_institution
    ON mdl_mst_nurse(institution_id, active);


-- ---- 20260425_add_trx_diagnosis.sql ----

-- Diagnosis transaction: one row per ICD-10 entry per visit
DO $$ BEGIN
    CREATE TYPE diagnosis_type AS ENUM ('primary', 'secondary', 'comorbidity');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE diagnosis_case AS ENUM ('new', 'chronic', 'acute_on_chronic');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE clinical_status_type AS ENUM ('active', 'recurrence', 'relapse', 'inactive', 'remission', 'resolved');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE verification_status_type AS ENUM ('unconfirmed', 'provisional', 'differential', 'confirmed', 'refuted', 'entered_in_error');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE prognosis_type AS ENUM ('sanam', 'bonam', 'dubia_ad_sanam', 'dubia_ad_malam', 'malam');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS mdl_trx_diagnosis (
    id                    BIGSERIAL                PRIMARY KEY,
    visit_id              BIGINT                   NOT NULL,
    institution_id        BIGINT                   NOT NULL,
    doctor_id             UUID                     NOT NULL,
    icd10_code            VARCHAR(10)              NOT NULL,
    icd10_display         TEXT                     NOT NULL,
    rank                  SMALLINT                 NOT NULL DEFAULT 1 CHECK (rank >= 1),
    type                  diagnosis_type           NOT NULL DEFAULT 'primary',
    "case"                diagnosis_case           NOT NULL DEFAULT 'new',
    clinical_status       clinical_status_type     NOT NULL DEFAULT 'active',
    verification_status   verification_status_type NOT NULL DEFAULT 'confirmed',
    prognosis             prognosis_type           NOT NULL DEFAULT 'malam',
    note                  TEXT,
    onset_date            DATE,
    -- SatuSehat
    satusehat_condition_id VARCHAR(100),
    -- Soft-delete
    deleted_at            TIMESTAMP,
    created_at            TIMESTAMP                DEFAULT NOW(),
    updated_at            TIMESTAMP                DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_trx_diagnosis_primary_rank'
          AND conrelid = 'mdl_trx_diagnosis'::regclass
    ) THEN
        ALTER TABLE mdl_trx_diagnosis
            ADD CONSTRAINT chk_trx_diagnosis_primary_rank
            CHECK (type <> 'primary' OR rank = 1);
    END IF;
END $$;

-- Partial index speeds up active-record queries (the common read path)
CREATE INDEX IF NOT EXISTS idx_trx_diagnosis_visit_active
    ON mdl_trx_diagnosis(institution_id, visit_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_trx_diagnosis_icd10
    ON mdl_trx_diagnosis(institution_id, icd10_code) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_trx_diagnosis_primary_unique_active
    ON mdl_trx_diagnosis (institution_id, visit_id)
    WHERE type = 'primary' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_trx_diagnosis_inst_id_active
    ON mdl_trx_diagnosis (institution_id, id)
    WHERE deleted_at IS NULL;


-- ---- 20260426_add_trx_anamnesa.sql ----

-- Anamnesa (subjective + vital signs + GCS) per visit — one active row per visit
DO $$ BEGIN
    CREATE TYPE respiratory_rate_unit_type AS ENUM ('breaths_per_minute');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE height_measurement_type AS ENUM ('berdiri', 'telentang');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE consciousness_type AS ENUM ('compos mentis', 'somnolen', 'sopor', 'coma');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE heart_rhythm_type AS ENUM ('regular', 'irregular');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE triage_type AS ENUM ('gawat darurat', 'darurat', 'tidak gawat darurat', 'meninggal');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS mdl_trx_anamnesa (
    id                  UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id            BIGINT   NOT NULL,
    institution_id      BIGINT   NOT NULL,
    nurse_id            UUID     NULL,
    doctor_id           UUID     NULL,
    -- Subjective
    chief_complaint     TEXT,
    secondary_complaint TEXT NULL,
    history_of_illness  TEXT,
    -- Illness duration
    illness_years       SMALLINT NOT NULL DEFAULT 0,
    illness_months      SMALLINT NOT NULL DEFAULT 0,
    illness_days        SMALLINT NOT NULL DEFAULT 0,
    -- Vital signs
    vs_systolic         SMALLINT,
    vs_diastolic        SMALLINT,
    vs_pulse            SMALLINT,
    vs_temperature      NUMERIC(4,1),
    vs_respiratory_rate SMALLINT,
    vs_oxygen_saturation SMALLINT,
    -- Derived (computed by usecase, stored for reporting)
    vs_map              SMALLINT,
    vs_weight           NUMERIC(5,1),
    vs_height           NUMERIC(5,1),
    vs_bmi              NUMERIC(5,2),
    vs_bmi_result       VARCHAR(30),
    vs_height_measurement      height_measurement_type,
    vs_abdominal_circumference NUMERIC(5,1),
    vs_consciousness           consciousness_type,
    vs_heart_rhythm            heart_rhythm_type,
    vs_pregnancy_status        BOOLEAN,
    vs_triage                  triage_type,
    -- GCS
    gcs_eye             SMALLINT,
    gcs_verbal          SMALLINT,
    gcs_motor           SMALLINT,
    -- gcs_total is a stored generated column so the DB keeps it consistent
    gcs_total           SMALLINT GENERATED ALWAYS AS (
                            COALESCE(gcs_eye, 0) + COALESCE(gcs_verbal, 0) + COALESCE(gcs_motor, 0)
                        ) STORED,
    -- Pain assessment
    pain_has_pain       BOOLEAN,
    pain_trigger        TEXT,
    pain_quality        VARCHAR(20) CHECK (pain_quality IN ('tekanan','terbakar','melilit','tertusuk','diiris','mencengkram')),
    pain_location       TEXT,
    pain_scale          SMALLINT CHECK (pain_scale BETWEEN 0 AND 10),
    pain_pattern        VARCHAR(20) CHECK (pain_pattern IN ('intermittent','continuous')),
    -- Timestamps
    created_at          TIMESTAMP DEFAULT NOW(),
    updated_at          TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_trx_anamnesa_visit
    ON mdl_trx_anamnesa(institution_id, visit_id);


-- ---- 20260427_add_satusehat_queue.sql ----

-- Transactional outbox queue for async SatuSehat FHIR submissions
DO $$ BEGIN
    CREATE TYPE satusehat_event_type AS ENUM ('diagnosis_save', 'anamnesa_save');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE satusehat_queue_status AS ENUM ('pending', 'processing', 'done', 'failed');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS mdl_trx_satusehat_queue (
    id              UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id        BIGINT                  NOT NULL,
    institution_id  BIGINT                  NOT NULL,
    event_type      satusehat_event_type    NOT NULL,
    payload         JSONB                   NOT NULL DEFAULT '{}',
    status          satusehat_queue_status  NOT NULL DEFAULT 'pending',
    attempts        SMALLINT                NOT NULL DEFAULT 0,
    last_error      TEXT,
    process_after   TIMESTAMP               DEFAULT NOW(),
    created_at      TIMESTAMP               DEFAULT NOW(),
    updated_at      TIMESTAMP               DEFAULT NOW()
);

-- Partial index — worker scans only pending rows ordered by process_after
CREATE INDEX IF NOT EXISTS idx_satusehat_queue_pending
    ON mdl_trx_satusehat_queue(process_after)
    WHERE status = 'pending';


-- ---- 20260429_add_illness_history_allergy_vitals_pain_to_trx_anamnesa.sql ----

-- Add illness duration, extended vital signs, and pain assessment columns
-- to mdl_trx_anamnesa.
-- fall_risk and lifestyle columns are intentionally omitted (not persisted).

DO $$ BEGIN CREATE TYPE height_measurement_type AS ENUM ('berdiri', 'telentang');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN CREATE TYPE consciousness_type AS ENUM ('compos mentis', 'somnolen', 'sopor', 'coma');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN CREATE TYPE heart_rhythm_type AS ENUM ('regular', 'irregular');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN CREATE TYPE triage_type AS ENUM ('gawat darurat', 'darurat', 'tidak gawat darurat', 'meninggal');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

ALTER TABLE mdl_trx_anamnesa
    -- Illness duration
    ADD COLUMN IF NOT EXISTS illness_years   SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS illness_months  SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS illness_days    SMALLINT NOT NULL DEFAULT 0,

    -- Extended vital signs
    ADD COLUMN IF NOT EXISTS vs_height_measurement      height_measurement_type NULL,
    ADD COLUMN IF NOT EXISTS vs_abdominal_circumference NUMERIC(5,1)            NULL,
    ADD COLUMN IF NOT EXISTS vs_consciousness           consciousness_type       NULL,
    ADD COLUMN IF NOT EXISTS vs_heart_rhythm            heart_rhythm_type        NULL,
    ADD COLUMN IF NOT EXISTS vs_pregnancy_status        BOOLEAN                  NULL,
    ADD COLUMN IF NOT EXISTS vs_triage                  triage_type              NULL,

    -- Pain assessment
    ADD COLUMN IF NOT EXISTS pain_has_pain  BOOLEAN     NULL,
    ADD COLUMN IF NOT EXISTS pain_trigger   TEXT        NULL,
    ADD COLUMN IF NOT EXISTS pain_quality   VARCHAR(20) NULL
        CHECK (pain_quality IN ('tekanan','terbakar','melilit','tertusuk','diiris','mencengkram')),
    ADD COLUMN IF NOT EXISTS pain_location  TEXT        NULL,
    ADD COLUMN IF NOT EXISTS pain_scale     SMALLINT    NULL CHECK (pain_scale BETWEEN 0 AND 10),
    ADD COLUMN IF NOT EXISTS pain_pattern   VARCHAR(20) NULL
        CHECK (pain_pattern IN ('intermittent','continuous'));


-- ---- 20260516_add_idx_visit_product_institution_time.sql ----

-- Speed up institution-scoped product order statistics queries.
CREATE INDEX IF NOT EXISTS idx_trx_visit_product_institution_create_time
    ON mdl_trx_visit_product (id_mst_institution, create_time)
    WHERE delete_time IS NULL;


-- ---- 20260611_add_rbac_permissions.sql ----

-- RBAC: permission master and role-permission mapping

CREATE TABLE IF NOT EXISTS public.mdl_mst_permission (
    id serial4 NOT NULL,
    code varchar NOT NULL,
    resource varchar NOT NULL,
    action varchar NOT NULL,
    description varchar NULL,
    create_time timestamptz NOT NULL DEFAULT now(),
    update_time timestamptz NOT NULL DEFAULT now(),
    delete_time timestamptz NULL,
    CONSTRAINT mdl_mst_permission_pkey PRIMARY KEY (id),
    CONSTRAINT mdl_mst_permission_code_key UNIQUE (code)
);

CREATE TABLE IF NOT EXISTS public.mdl_map_role_permission (
    id serial4 NOT NULL,
    id_mst_role int8 NOT NULL,
    id_mst_permission int8 NOT NULL,
    CONSTRAINT mdl_map_role_permission_pkey PRIMARY KEY (id),
    CONSTRAINT mdl_map_role_permission_role_permission_key UNIQUE (id_mst_role, id_mst_permission)
);

CREATE INDEX IF NOT EXISTS idx_mdl_map_role_permission_role
    ON public.mdl_map_role_permission (id_mst_role);

CREATE INDEX IF NOT EXISTS idx_mdl_map_role_permission_permission
    ON public.mdl_map_role_permission (id_mst_permission);

-- Seed permissions (codes must match internal/entity/constant/permission/permission.go)
INSERT INTO public.mdl_mst_permission (code, resource, action, description)
SELECT v.code, v.resource, v.action, v.description
FROM (VALUES
    ('patient.read', 'patient', 'read', 'View patients'),
    ('patient.create', 'patient', 'create', 'Register patients'),
    ('patient.update', 'patient', 'update', 'Update patients'),
    ('visit.read', 'visit', 'read', 'View visits'),
    ('visit.create', 'visit', 'create', 'Create visits'),
    ('visit.update', 'visit', 'update', 'Update visits'),
    ('diagnosis.read', 'diagnosis', 'read', 'View diagnoses'),
    ('diagnosis.create', 'diagnosis', 'create', 'Create diagnoses'),
    ('diagnosis.update', 'diagnosis', 'update', 'Update diagnoses'),
    ('diagnosis.delete', 'diagnosis', 'delete', 'Delete diagnoses'),
    ('anamnesa.read', 'anamnesa', 'read', 'View anamnesa'),
    ('anamnesa.create', 'anamnesa', 'create', 'Create anamnesa'),
    ('anamnesa.update', 'anamnesa', 'update', 'Update anamnesa'),
    ('anamnesa.delete', 'anamnesa', 'delete', 'Delete anamnesa'),
    ('product.read', 'product', 'read', 'View products'),
    ('product.create', 'product', 'create', 'Create products'),
    ('product.update', 'product', 'update', 'Update products'),
    ('product.delete', 'product', 'delete', 'Delete products'),
    ('product.statistics', 'product', 'statistics', 'View product metrics and statistics'),
    ('journey.read', 'journey', 'read', 'View journey boards and points'),
    ('journey.create', 'journey', 'create', 'Create journey boards and points'),
    ('journey.update', 'journey', 'update', 'Update journey boards and points'),
    ('journey.delete', 'journey', 'delete', 'Delete journey boards and points'),
    ('recall.read', 'recall', 'read', 'View recalls'),
    ('recall.create', 'recall', 'create', 'Create recalls'),
    ('recall.update', 'recall', 'update', 'Update recalls'),
    ('recall.delete', 'recall', 'delete', 'Delete recalls'),
    ('odontogram.read', 'odontogram', 'read', 'View odontogram'),
    ('odontogram.create', 'odontogram', 'create', 'Create odontogram events'),
    ('odontogram.update', 'odontogram', 'update', 'Update odontogram'),
    ('odontogram.delete', 'odontogram', 'delete', 'Delete odontogram events'),
    ('reference.search', 'reference', 'search', 'Search ICD-10, doctors, and nurses')
) AS v(code, resource, action, description)
WHERE NOT EXISTS (
    SELECT 1 FROM public.mdl_mst_permission p WHERE p.code = v.code
);

-- Administrator: all permissions
INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
CROSS JOIN public.mdl_mst_permission p
WHERE r.name = 'administrator'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );

-- Clerk: patient, visit, recall, reference
INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code IN (
    'patient.read', 'patient.create', 'patient.update',
    'visit.read', 'visit.create', 'visit.update',
    'recall.read', 'recall.create', 'recall.update', 'recall.delete',
    'product.statistics',
    'reference.search'
)
WHERE r.name = 'clerk'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );

-- Doctor: patient.read, visit.*, diagnosis.*, anamnesa.*, odontogram.*, reference.search
INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code IN (
    'patient.read',
    'visit.read', 'visit.create', 'visit.update',
    'diagnosis.read', 'diagnosis.create', 'diagnosis.update', 'diagnosis.delete',
    'anamnesa.read', 'anamnesa.create', 'anamnesa.update', 'anamnesa.delete',
    'odontogram.read', 'odontogram.create', 'odontogram.update', 'odontogram.delete',
    'reference.search'
)
WHERE r.name = 'doctor'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );

-- Nurse: patient.read, visit.read, anamnesa.read, odontogram.*, recall.*, reference.search
INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code IN (
    'patient.read',
    'visit.read',
    'anamnesa.read',
    'odontogram.read', 'odontogram.create', 'odontogram.update', 'odontogram.delete',
    'recall.read', 'recall.create', 'recall.update', 'recall.delete',
    'reference.search'
)
WHERE r.name = 'nurse'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );


-- ---- 20260614_enforce_unique_active_staff_email.sql ----

-- Enforce global uniqueness of active staff emails for deterministic SSO login
-- (soft-deleted staff can reuse old emails after reactivation policy decisions).

CREATE UNIQUE INDEX IF NOT EXISTS uidx_mdl_mst_staff_active_email
    ON public.mdl_mst_staff (LOWER(email))
    WHERE delete_time IS NULL;


-- ---- 20260706_add_ref_icd9cm.sql ----

-- ICD-9-CM procedure code reference table (Indonesian healthcare standard).
-- Seeded from docs/ICD9 CM.csv via scripts/icd9cm_csv_to_sql.py.
-- Global reference data — not institution-scoped.
CREATE TABLE IF NOT EXISTS mdl_ref_icd9cm (
    code        VARCHAR(10)  PRIMARY KEY,
    display     VARCHAR(500) NOT NULL,
    parent_code VARCHAR(10),
    depth       SMALLINT     NOT NULL,
    is_leaf     BOOLEAN      NOT NULL DEFAULT FALSE,
    version     VARCHAR(20)  NOT NULL DEFAULT 'ICD9CM_2010'
);

CREATE INDEX IF NOT EXISTS idx_mdl_ref_icd9cm_parent_code
    ON mdl_ref_icd9cm (parent_code);

CREATE INDEX IF NOT EXISTS idx_mdl_ref_icd9cm_search
    ON mdl_ref_icd9cm USING gin (to_tsvector('simple', code || ' ' || display));


-- ---- 20260707_add_procedure_tables.sql ----

-- Visit procedure transaction: one row per procedure recorded in a visit.
-- product_id optionally references mdl_trx_institution_product (is_treatment = true).
-- Snapshot columns (product_name, doctor_name, nurse_name, icd10pcs_display)
-- are captured at write time so reads never need JOINs and history is preserved
-- even if master data changes.
CREATE TABLE IF NOT EXISTS mdl_trx_visit_procedure (
    id                  BIGSERIAL       PRIMARY KEY,
    visit_id            BIGINT          NOT NULL,
    institution_id      BIGINT          NOT NULL,
    product_id          BIGINT,
    product_name        VARCHAR(255),
    doctor_id           VARCHAR(50)     NOT NULL,
    doctor_name         VARCHAR(255)    NOT NULL,
    nurse_id            VARCHAR(50),
    nurse_name          VARCHAR(255),
    planned_at          TIMESTAMPTZ,
    category            VARCHAR(20),
    duration            VARCHAR(100),
    icd9cm_code         VARCHAR(10),
    icd9cm_display      VARCHAR(500),
    description         TEXT,
    notes               TEXT,
    rank                SMALLINT        NOT NULL CHECK (rank >= 1),
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_trx_visit_procedure_visit_active
    ON mdl_trx_visit_procedure (institution_id, visit_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_trx_visit_procedure_inst_id_active
    ON mdl_trx_visit_procedure (institution_id, id)
    WHERE deleted_at IS NULL;


-- ---- 20260804_add_idx_recall_visit.sql ----

CREATE INDEX IF NOT EXISTS idx_trx_recall_visit
    ON public.mdl_trx_recall(id_trx_patient_visit)
    WHERE delete_time IS NULL AND id_trx_patient_visit IS NOT NULL;


-- ---- 20260810_add_idx_patient_name_fts.sql ----

CREATE INDEX IF NOT EXISTS idx_mst_patient_institution_name
    ON mdl_mst_patient_institution USING gin (to_tsvector('simple', name));


-- ---- 20260901_add_compensation_tables.sql ----

-- Compensation domain: staff wages, payday periods, visit commissions,
-- manual visit contributors, and visit lock columns.
--
-- ID strategy (PRD/TRD + live schema):
--   institution_id, visit_id → BIGINT (mdl_mst_institution.id, mdl_trx_patient_visit.id)
--   staff_id / audit staff cols → UUID (mdl_mst_staff.uuid; same as doctor/nurse staff_uuid)
--   mdl_trx_compensation_period.uuid → public API key only
--
-- No physical REFERENCES / ON DELETE (project convention).

-- ---------------------------------------------------------------------------
-- Enums
-- ---------------------------------------------------------------------------
DO $$ BEGIN
    CREATE TYPE wage_cadence_enum AS ENUM ('monthly', 'weekly');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE compensation_period_status_enum AS ENUM ('open', 'draft', 'finalized');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE commission_type_enum AS ENUM ('percent', 'flat');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ---------------------------------------------------------------------------
-- mdl_mst_staff_wage
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mdl_mst_staff_wage (
    id              BIGSERIAL           PRIMARY KEY,
    staff_id        UUID                NOT NULL,
    institution_id  BIGINT              NOT NULL,
    wage_amount     BIGINT              NOT NULL,
    wage_cadence    wage_cadence_enum   NOT NULL,
    is_active       BOOLEAN             NOT NULL DEFAULT true,
    effective_from  DATE                NOT NULL,
    effective_to    DATE,
    created_by      UUID,
    updated_by      UUID,
    create_time     TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    update_time     TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    delete_time     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_staff_wage_staff_institution
    ON mdl_mst_staff_wage (staff_id, institution_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_staff_wage_effective
    ON mdl_mst_staff_wage (staff_id, institution_id, effective_from, effective_to)
    WHERE delete_time IS NULL;

-- ---------------------------------------------------------------------------
-- mdl_trx_compensation_period
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mdl_trx_compensation_period (
    id              BIGSERIAL                           PRIMARY KEY,
    uuid            UUID                                NOT NULL DEFAULT gen_random_uuid(),
    institution_id  BIGINT                              NOT NULL,
    label           VARCHAR(100)                        NOT NULL,
    period_start    DATE                                NOT NULL,
    period_end      DATE                                NOT NULL,
    status          compensation_period_status_enum     NOT NULL DEFAULT 'open',
    wage_snapshot   JSONB,
    total_wage      BIGINT,
    total_commission BIGINT,
    total_payout    BIGINT,
    staff_count     INT,
    visit_count     INT,
    drafted_at      TIMESTAMPTZ,
    drafted_by      UUID,
    finalized_at    TIMESTAMPTZ,
    finalized_by    UUID,
    create_time     TIMESTAMPTZ                         NOT NULL DEFAULT NOW(),
    update_time     TIMESTAMPTZ                         NOT NULL DEFAULT NOW(),
    delete_time     TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_comp_period_uuid
    ON mdl_trx_compensation_period (uuid);

CREATE INDEX IF NOT EXISTS idx_comp_period_institution_status
    ON mdl_trx_compensation_period (institution_id, status)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_comp_period_dates
    ON mdl_trx_compensation_period (institution_id, period_start, period_end)
    WHERE delete_time IS NULL;

-- ---------------------------------------------------------------------------
-- mdl_trx_worksheet — single-staff wrap for visit commissions
-- ---------------------------------------------------------------------------
DO $$ BEGIN
    CREATE TYPE worksheet_status_enum AS ENUM ('pending', 'open', 'finalized');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE worksheet_generate_status_enum AS ENUM ('idle', 'running', 'succeeded', 'failed');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS mdl_trx_worksheet (
    id                      BIGSERIAL                           PRIMARY KEY,
    uuid                    UUID                                NOT NULL DEFAULT gen_random_uuid(),
    institution_id          BIGINT                              NOT NULL,
    staff_id                UUID                                NOT NULL,
    label                   VARCHAR(100)                        NOT NULL,
    period_start            DATE                                NOT NULL,
    period_end              DATE                                NOT NULL,
    status                  worksheet_status_enum               NOT NULL DEFAULT 'open',
    generate_status         worksheet_generate_status_enum      NOT NULL DEFAULT 'idle',
    compensation_period_id  BIGINT,
    total_commission        BIGINT,
    visit_count             INT,
    generate_started_at     TIMESTAMPTZ,
    generate_finished_at    TIMESTAMPTZ,
    generate_error          TEXT,
    finalized_at            TIMESTAMPTZ,
    finalized_by            UUID,
    created_by              UUID,
    create_time             TIMESTAMPTZ                         NOT NULL DEFAULT NOW(),
    update_time             TIMESTAMPTZ                         NOT NULL DEFAULT NOW(),
    delete_time             TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_worksheet_uuid
    ON mdl_trx_worksheet (uuid);

CREATE INDEX IF NOT EXISTS idx_worksheet_institution_staff_status
    ON mdl_trx_worksheet (institution_id, staff_id, status)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_worksheet_institution_staff_dates
    ON mdl_trx_worksheet (institution_id, staff_id, period_start, period_end)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_worksheet_compensation_period
    ON mdl_trx_worksheet (compensation_period_id)
    WHERE delete_time IS NULL AND compensation_period_id IS NOT NULL;

-- ---------------------------------------------------------------------------
-- mdl_trx_visit_commission
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mdl_trx_visit_commission (
    id                      BIGSERIAL               PRIMARY KEY,
    worksheet_id            BIGINT                  NOT NULL,
    visit_id                BIGINT                  NOT NULL,
    staff_id                UUID                    NOT NULL,
    revenue_base            BIGINT                  NOT NULL,
    commission_type         commission_type_enum    NOT NULL,
    commission_percent      DECIMAL(5,2),
    commission_flat_amount  BIGINT,
    commission_amount       BIGINT                  NOT NULL,
    sources                 JSONB,
    note                    TEXT,
    included_manually       BOOLEAN                 NOT NULL DEFAULT false,
    create_time             TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    update_time             TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    delete_time             TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_commission_worksheet_visit
    ON mdl_trx_visit_commission (worksheet_id, visit_id)
    WHERE delete_time IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uidx_commission_visit_staff
    ON mdl_trx_visit_commission (visit_id, staff_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_commission_worksheet
    ON mdl_trx_visit_commission (worksheet_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_commission_visit
    ON mdl_trx_visit_commission (visit_id)
    WHERE delete_time IS NULL;

-- ---------------------------------------------------------------------------
-- mdl_map_visit_contributor
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS mdl_map_visit_contributor (
    id              BIGSERIAL       PRIMARY KEY,
    visit_id        BIGINT          NOT NULL,
    staff_id        UUID            NOT NULL,
    institution_id  BIGINT          NOT NULL,
    added_by        UUID            NOT NULL,
    create_time     TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    delete_time     TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_contributor_visit_staff
    ON mdl_map_visit_contributor (visit_id, staff_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_contributor_visit
    ON mdl_map_visit_contributor (visit_id)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_contributor_staff
    ON mdl_map_visit_contributor (staff_id, institution_id)
    WHERE delete_time IS NULL;

-- ---------------------------------------------------------------------------
-- Alter mdl_trx_patient_visit — compensation lock columns
-- ---------------------------------------------------------------------------
ALTER TABLE mdl_trx_patient_visit
    ADD COLUMN IF NOT EXISTS compensation_period_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS compensation_locked_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS worksheet_id BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_visit_compensation_lock
    ON mdl_trx_patient_visit (compensation_locked_at)
    WHERE compensation_locked_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_visit_worksheet_lock
    ON mdl_trx_patient_visit (worksheet_id)
    WHERE worksheet_id IS NOT NULL;


-- ---- 20260902_add_idx_visit_product_visit.sql ----

-- Speed up full-visit revenue_base aggregation for compensation (Option A).
CREATE INDEX IF NOT EXISTS idx_trx_visit_product_visit
    ON mdl_trx_visit_product (id_trx_patient_visit)
    WHERE delete_time IS NULL;


-- ---- 20260907_add_idx_mst_staff_uuid.sql ----

-- Name lookup and visit-contributor map joins use mdl_mst_staff.uuid
-- (varchar, PK is numeric id). Unique among live rows.
CREATE UNIQUE INDEX IF NOT EXISTS uidx_mst_staff_uuid
    ON mdl_mst_staff (uuid)
    WHERE delete_time IS NULL;


-- ---- 20260907_add_idx_visit_institution_create_time.sql ----

-- Period staff-list detection bounds visits by institution + create_time.
CREATE INDEX IF NOT EXISTS idx_trx_patient_visit_institution_create_time
    ON mdl_trx_patient_visit (id_mst_institution, create_time)
    WHERE delete_time IS NULL;


-- ---- 20260908_procedure_doctor_nurse_uuid.sql ----

-- Align procedure practitioner FKs with mdl_mst_doctor.id / mdl_mst_nurse.id (UUID).
-- Enables index-friendly UUID equality joins in contributor detection SQL.
-- Idempotent: skips validate/cast when doctor_id / nurse_id are already uuid.
--
-- Pre-check (manual, varchar columns only): list invalid rows before applying if migration raises.
--   SELECT id, doctor_id, nurse_id
--   FROM mdl_trx_visit_procedure
--   WHERE doctor_id IS NULL
--      OR doctor_id::text !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
--      OR (nurse_id IS NOT NULL AND nurse_id::text <> ''
--          AND nurse_id::text !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$');

DO $$
DECLARE
    doctor_type text;
    nurse_type  text;
    bad_doctor  bigint;
    bad_nurse   bigint;
BEGIN
    SELECT data_type INTO doctor_type
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'mdl_trx_visit_procedure'
      AND column_name = 'doctor_id';

    SELECT data_type INTO nurse_type
    FROM information_schema.columns
    WHERE table_schema = current_schema()
      AND table_name = 'mdl_trx_visit_procedure'
      AND column_name = 'nurse_id';

    IF doctor_type IN ('character varying', 'character', 'text') THEN
        SELECT COUNT(*) INTO bad_doctor
        FROM mdl_trx_visit_procedure
        WHERE doctor_id IS NULL
           OR doctor_id::text !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

        IF bad_doctor > 0 THEN
            RAISE EXCEPTION
                'mdl_trx_visit_procedure: % row(s) have invalid doctor_id (not a UUID); fix before casting',
                bad_doctor;
        END IF;

        ALTER TABLE mdl_trx_visit_procedure
            ALTER COLUMN doctor_id TYPE UUID USING doctor_id::uuid;
    END IF;

    IF nurse_type IN ('character varying', 'character', 'text') THEN
        SELECT COUNT(*) INTO bad_nurse
        FROM mdl_trx_visit_procedure
        WHERE nurse_id IS NOT NULL
          AND nurse_id::text <> ''
          AND nurse_id::text !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

        IF bad_nurse > 0 THEN
            RAISE EXCEPTION
                'mdl_trx_visit_procedure: % row(s) have invalid nurse_id (not a UUID); fix before casting',
                bad_nurse;
        END IF;

        ALTER TABLE mdl_trx_visit_procedure
            ALTER COLUMN nurse_id TYPE UUID USING NULLIF(nurse_id, '')::uuid;
    END IF;
END $$;

-- Period staff detection filters contributors by institution then intersects period visits.
CREATE INDEX IF NOT EXISTS idx_contributor_institution_visit
    ON mdl_map_visit_contributor (institution_id, visit_id)
    WHERE delete_time IS NULL;


-- ---- 20260909_add_visit_commission_approved_at.sql ----

-- Membership rows from generate are unapproved until PUT sets approved_at.
ALTER TABLE mdl_trx_visit_commission
    ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_commission_worksheet_unapproved
    ON mdl_trx_visit_commission (worksheet_id)
    WHERE delete_time IS NULL AND approved_at IS NULL;

-- ---- 20260917120000_add_worksheet.sql ----
-- (desired state already reflected above: mdl_trx_worksheet + commission.worksheet_id)

