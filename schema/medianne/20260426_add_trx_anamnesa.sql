-- Anamnesa (subjective + vital signs + GCS) per visit — one active row per visit
DO $$ BEGIN
    CREATE TYPE respiratory_rate_unit_type AS ENUM ('breaths_per_minute');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS mdl_trx_anamnesa (
    id                  UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
    visit_id            BIGINT   NOT NULL,
    institution_id      BIGINT   NOT NULL,
    nurse_id            UUID     NULL,
    -- Subjective
    chief_complaint     TEXT,
    history_of_illness  TEXT,
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
    -- GCS
    gcs_eye             SMALLINT,
    gcs_verbal          SMALLINT,
    gcs_motor           SMALLINT,
    -- gcs_total is a stored generated column so the DB keeps it consistent
    gcs_total           SMALLINT GENERATED ALWAYS AS (
                            COALESCE(gcs_eye, 0) + COALESCE(gcs_verbal, 0) + COALESCE(gcs_motor, 0)
                        ) STORED,
    -- Timestamps
    created_at          TIMESTAMP DEFAULT NOW(),
    updated_at          TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_trx_anamnesa_visit
    ON mdl_trx_anamnesa(institution_id, visit_id);
