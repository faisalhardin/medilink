-- Staff management permissions

INSERT INTO public.mdl_mst_permission (code, resource, action, description)
SELECT v.code, v.resource, v.action, v.description
FROM (VALUES
    ('staff.read', 'staff', 'read', 'View staff members'),
    ('staff.create', 'staff', 'create', 'Create staff members'),
    ('staff.update', 'staff', 'update', 'Update staff information'),
    ('staff.delete', 'staff', 'delete', 'Activate/deactivate staff'),
    ('staff.role.assign', 'staff.role', 'assign', 'Assign and unassign roles to staff')
) AS v(code, resource, action, description)
WHERE NOT EXISTS (
    SELECT 1 FROM public.mdl_mst_permission p WHERE p.code = v.code
);

INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code IN (
    'staff.read',
    'staff.create',
    'staff.update',
    'staff.delete',
    'staff.role.assign'
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
