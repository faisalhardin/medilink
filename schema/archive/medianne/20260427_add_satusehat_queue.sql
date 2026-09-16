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
