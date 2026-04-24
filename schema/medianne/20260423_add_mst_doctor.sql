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
