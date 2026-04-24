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
    CREATE TYPE prognosis_type AS ENUM ('excellent', 'good', 'fair', 'poor', 'unknown');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS mdl_trx_diagnosis (
    id                    UUID                     PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id              BIGINT                   NOT NULL,
    institution_id        BIGINT                   NOT NULL,
    doctor_id             UUID                     NOT NULL,
    icd10_code            VARCHAR(10)              NOT NULL,
    type                  diagnosis_type           NOT NULL DEFAULT 'primary',
    case                  diagnosis_case           NOT NULL DEFAULT 'new',
    clinical_status       clinical_status_type     NOT NULL DEFAULT 'active',
    verification_status   verification_status_type NOT NULL DEFAULT 'confirmed',
    prognosis             prognosis_type           NOT NULL DEFAULT 'unknown',
    note                  TEXT,
    onset_date            DATE,
    -- SatuSehat
    satusehat_condition_id VARCHAR(100),
    -- Soft-delete
    deleted_at            TIMESTAMP,
    created_at            TIMESTAMP                DEFAULT NOW(),
    updated_at            TIMESTAMP                DEFAULT NOW()
);

-- Partial index speeds up active-record queries (the common read path)
CREATE INDEX IF NOT EXISTS idx_trx_diagnosis_visit_active
    ON mdl_trx_diagnosis(institution_id, visit_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_trx_diagnosis_icd10
    ON mdl_trx_diagnosis(institution_id, icd10_code) WHERE deleted_at IS NULL;
