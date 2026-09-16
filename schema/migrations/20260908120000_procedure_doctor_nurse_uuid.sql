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
