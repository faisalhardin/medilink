-- One row per room and column. A column may keep zero rooms.
-- Delete null ids first so the primary key can be added.
-- Keep the lowest ctid when a pair was stored more than once.

DELETE FROM mdl_map_service_point_journey_point
WHERE id_mst_service_point IS NULL
   OR id_mst_journey_point IS NULL;

DELETE FROM mdl_map_service_point_journey_point AS extra
WHERE extra.ctid IN (
    SELECT ctid
    FROM (
        SELECT ctid,
               row_number() OVER (
                   PARTITION BY id_mst_service_point, id_mst_journey_point
                   ORDER BY ctid
               ) AS n
        FROM mdl_map_service_point_journey_point
    ) ranked
    WHERE ranked.n > 1
);

ALTER TABLE mdl_map_service_point_journey_point
    ADD PRIMARY KEY (id_mst_service_point, id_mst_journey_point);

CREATE INDEX idx_map_service_point_journey_point_journey_point
    ON mdl_map_service_point_journey_point (id_mst_journey_point);
