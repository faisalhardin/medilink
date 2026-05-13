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
