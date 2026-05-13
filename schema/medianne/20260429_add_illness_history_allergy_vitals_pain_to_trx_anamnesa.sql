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
