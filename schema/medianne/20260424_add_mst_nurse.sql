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
