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
