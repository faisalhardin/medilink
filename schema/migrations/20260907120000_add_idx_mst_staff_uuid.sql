-- Name lookup and visit-contributor map joins use mdl_mst_staff.uuid
-- (varchar, PK is numeric id). Unique among live rows.
CREATE UNIQUE INDEX IF NOT EXISTS uidx_mst_staff_uuid
    ON mdl_mst_staff (uuid)
    WHERE delete_time IS NULL;
