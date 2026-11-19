-- One live room name per clinic (BE-1).
-- Order is function, then backfill, then guards, then CHECK, then indexes.
-- Atlas runs this file in one transaction. A RAISE rolls the backfill back.
-- Re-apply is safe because the backfill only touches clinic-0 rows.
--
-- Leftover report (run after apply; Atlas does not print RAISE NOTICE):
--   SELECT sp.id, sp.name, sp.id_mst_journey_board, sp.delete_time,
--          CASE
--              WHEN b.id IS NULL THEN 'board missing'
--              WHEN b.id_mst_institution = 0 THEN 'board institution 0'
--          END AS board_reason,
--          (sp.delete_time IS NOT NULL) AS archived
--   FROM mdl_mst_service_point sp
--   LEFT JOIN mdl_mst_journey_board b ON b.id = sp.id_mst_journey_board
--   WHERE sp.id_mst_institution = 0
--   ORDER BY sp.id;
--
-- Duplicate pre-check (same grouping the guard uses):
--   SELECT id_mst_institution, mdl_service_point_name_key(name), array_agg(id ORDER BY id)
--   FROM mdl_mst_service_point
--   WHERE delete_time IS NULL AND id_mst_institution <> 0
--   GROUP BY 1, 2 HAVING count(*) > 1;

CREATE OR REPLACE FUNCTION mdl_service_point_name_key(name text)
RETURNS text
LANGUAGE SQL
IMMUTABLE
PARALLEL SAFE
AS $$
    SELECT lower(btrim(name))
$$;

DO $$
DECLARE
    n bigint;
    leftover_count bigint;
    rec record;
    blank_list text;
    dup_groups bigint;
    dup_list text;
BEGIN
    UPDATE mdl_mst_service_point AS sp
    SET id_mst_institution = b.id_mst_institution,
        update_time = now()
    FROM mdl_mst_journey_board AS b
    WHERE sp.id_mst_institution = 0
      AND sp.id_mst_journey_board = b.id
      AND b.id_mst_institution <> 0;

    GET DIAGNOSTICS n = ROW_COUNT;
    RAISE NOTICE 'mdl_mst_service_point institution backfill from board: %', n;

    SELECT count(*) INTO leftover_count
    FROM mdl_mst_service_point
    WHERE id_mst_institution = 0;

    RAISE NOTICE 'mdl_mst_service_point: % room(s) left at institution 0', leftover_count;

    FOR rec IN
        SELECT sp.id,
               sp.name,
               sp.id_mst_journey_board AS board_id,
               CASE
                   WHEN b.id IS NULL THEN 'board missing'
                   WHEN b.id_mst_institution = 0 THEN 'board institution 0'
               END AS board_reason,
               (sp.delete_time IS NOT NULL) AS archived
        FROM mdl_mst_service_point AS sp
        LEFT JOIN mdl_mst_journey_board AS b
          ON b.id = sp.id_mst_journey_board
        WHERE sp.id_mst_institution = 0
        ORDER BY sp.id
    LOOP
        RAISE NOTICE '%', format(
            'mdl_mst_service_point leftover institution 0 id=%s name=%L board=%s reason=%s archived=%s',
            rec.id, rec.name, rec.board_id, rec.board_reason, rec.archived
        );
    END LOOP;

    SELECT string_agg(
               format('id=%s institution=%s name=%L', id, id_mst_institution, name),
               E'\n' ORDER BY id
           )
      INTO blank_list
      FROM mdl_mst_service_point
     WHERE mdl_service_point_name_key(name) = '';

    IF blank_list IS NOT NULL THEN
        RAISE EXCEPTION E'mdl_mst_service_point: blank room names; rename by hand, then re-apply:\n%',
            blank_list;
    END IF;

    -- Predicate must equal ux_service_point_institution_name's WHERE clause.
    SELECT count(*),
           string_agg(
               format(
                   'institution=%s key=%L ids=%s names=%s',
                   id_mst_institution, name_key, ids, names
               ),
               E'\n' ORDER BY id_mst_institution, name_key
           )
      INTO dup_groups, dup_list
      FROM (
            SELECT id_mst_institution,
                   mdl_service_point_name_key(name) AS name_key,
                   string_agg(id::text, ',' ORDER BY id) AS ids,
                   string_agg(quote_literal(name), ', ' ORDER BY id) AS names
              FROM mdl_mst_service_point
             WHERE delete_time IS NULL
               AND id_mst_institution <> 0
             GROUP BY 1, 2
            HAVING count(*) > 1
           ) d;

    IF dup_groups > 0 THEN
        RAISE EXCEPTION E'mdl_mst_service_point: % duplicate live room name group(s) per clinic; rename by hand, then re-apply:\n%',
            dup_groups, dup_list;
    END IF;
END $$;

ALTER TABLE mdl_mst_service_point
    ADD CONSTRAINT mdl_mst_service_point_name_not_blank
    CHECK (mdl_service_point_name_key(name) <> '');

CREATE UNIQUE INDEX IF NOT EXISTS ux_service_point_institution_name
    ON mdl_mst_service_point (id_mst_institution, mdl_service_point_name_key(name))
    WHERE delete_time IS NULL AND id_mst_institution <> 0;

CREATE INDEX IF NOT EXISTS idx_service_point_institution_board
    ON mdl_mst_service_point (id_mst_institution, id_mst_journey_board)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_service_point_board
    ON mdl_mst_service_point (id_mst_journey_board)
    WHERE delete_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_patient_visit_service_point
    ON mdl_trx_patient_visit (id_mst_service_point, id_mst_institution)
    WHERE delete_time IS NULL AND id_mst_service_point IS NOT NULL;
