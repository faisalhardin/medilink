-- Backfill mdl_mst_doctor / mdl_mst_nurse from existing staff + roles.
--
-- Source tables (legacy auth schema):
--   - mdl_mst_staff (uuid is stored as varchar)
--   - mdl_map_role_staff (maps staff.id -> role.id)
--   - mdl_mst_role (name includes 'doctor' and 'nurse')
--
-- Target tables (new practitioner masters):
--   - mdl_mst_doctor (staff_uuid UUID UNIQUE)
--   - mdl_mst_nurse  (staff_uuid UUID UNIQUE)
--
-- Idempotency:
--   - Uses NOT EXISTS checks on target.staff_uuid so it is safe to re-run.
--
-- Safety:
--   - Filters out soft-deleted staff / roles (delete_time IS NULL).
--   - Skips staff rows whose uuid is not a valid UUID string (to avoid ::uuid cast failure).
--
-- NOTE:
--   This migration only handles role names 'doctor' and 'nurse' as requested.

-- Insert doctors for staff with role 'doctor'
INSERT INTO mdl_mst_doctor (staff_uuid, name, institution_id, active, created_at, updated_at)
SELECT DISTINCT
    s."uuid"::uuid        AS staff_uuid,
    s."name"              AS name,
    s.id_mst_institution  AS institution_id,
    TRUE                  AS active,
    NOW()                 AS created_at,
    NOW()                 AS updated_at
FROM mdl_mst_staff s
JOIN mdl_map_role_staff mrs ON mrs.id_mst_staff = s.id
JOIN mdl_mst_role r         ON r.id = mrs.id_mst_role AND r.delete_time IS NULL
WHERE s.delete_time IS NULL
  AND r.name = 'doctor'
  AND s."uuid" ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
  AND NOT EXISTS (
      SELECT 1
      FROM mdl_mst_doctor d
      WHERE d.staff_uuid = s."uuid"::uuid
  );

-- Insert nurses for staff with role 'nurse'
INSERT INTO mdl_mst_nurse (staff_uuid, name, role, institution_id, active, created_at, updated_at)
SELECT DISTINCT
    s."uuid"::uuid        AS staff_uuid,
    s."name"              AS name,
    'nurse'::nurse_role_type AS role,
    s.id_mst_institution  AS institution_id,
    TRUE                  AS active,
    NOW()                 AS created_at,
    NOW()                 AS updated_at
FROM mdl_mst_staff s
JOIN mdl_map_role_staff mrs ON mrs.id_mst_staff = s.id
JOIN mdl_mst_role r         ON r.id = mrs.id_mst_role AND r.delete_time IS NULL
WHERE s.delete_time IS NULL
  AND r.name = 'nurse'
  AND s."uuid" ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
  AND NOT EXISTS (
      SELECT 1
      FROM mdl_mst_nurse n
      WHERE n.staff_uuid = s."uuid"::uuid
  );

