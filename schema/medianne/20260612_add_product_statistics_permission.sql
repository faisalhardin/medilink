-- Add product.statistics permission for environments that already ran 20260611_add_rbac_permissions.sql

INSERT INTO public.mdl_mst_permission (code, resource, action, description)
SELECT 'product.statistics', 'product', 'statistics', 'View product metrics and statistics'
WHERE NOT EXISTS (
    SELECT 1 FROM public.mdl_mst_permission p WHERE p.code = 'product.statistics'
);

INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code = 'product.statistics'
WHERE r.name IN ('administrator', 'clerk')
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );
