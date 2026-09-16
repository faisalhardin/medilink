-- Compensation permissions (administrator only)

INSERT INTO public.mdl_mst_permission (code, resource, action, description)
SELECT v.code, v.resource, v.action, v.description
FROM (VALUES
    ('compensation.read', 'compensation', 'read', 'View payday list & summaries'),
    ('compensation.assign', 'compensation', 'assign', 'Set commission (% or flat), store draft'),
    ('compensation.finalize', 'compensation', 'finalize', 'Finalize period / unlock (admin)'),
    ('compensation.manage', 'compensation', 'manage', 'Monthly wage config')
) AS v(code, resource, action, description)
WHERE NOT EXISTS (
    SELECT 1 FROM public.mdl_mst_permission p WHERE p.code = v.code
);

INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code IN (
    'compensation.read',
    'compensation.assign',
    'compensation.finalize',
    'compensation.manage'
)
WHERE r.name = 'administrator'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );
