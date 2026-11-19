-- Service point manage permission (administrator only)

INSERT INTO public.mdl_mst_permission (code, resource, action, description)
SELECT v.code, v.resource, v.action, v.description
FROM (VALUES
    ('service_point.manage', 'service_point', 'manage', 'Manage clinic rooms (service points)')
) AS v(code, resource, action, description)
WHERE NOT EXISTS (
    SELECT 1 FROM public.mdl_mst_permission p WHERE p.code = v.code
);

INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code = 'service_point.manage'
WHERE r.name = 'administrator'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );
