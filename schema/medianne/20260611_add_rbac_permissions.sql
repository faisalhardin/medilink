-- RBAC: permission master and role-permission mapping

CREATE TABLE IF NOT EXISTS public.mdl_mst_permission (
    id serial4 NOT NULL,
    code varchar NOT NULL,
    resource varchar NOT NULL,
    action varchar NOT NULL,
    description varchar NULL,
    create_time timestamptz NOT NULL DEFAULT now(),
    update_time timestamptz NOT NULL DEFAULT now(),
    delete_time timestamptz NULL,
    CONSTRAINT mdl_mst_permission_pkey PRIMARY KEY (id),
    CONSTRAINT mdl_mst_permission_code_key UNIQUE (code)
);

CREATE TABLE IF NOT EXISTS public.mdl_map_role_permission (
    id serial4 NOT NULL,
    id_mst_role int8 NOT NULL,
    id_mst_permission int8 NOT NULL,
    CONSTRAINT mdl_map_role_permission_pkey PRIMARY KEY (id),
    CONSTRAINT mdl_map_role_permission_role_permission_key UNIQUE (id_mst_role, id_mst_permission)
);

CREATE INDEX IF NOT EXISTS idx_mdl_map_role_permission_role
    ON public.mdl_map_role_permission (id_mst_role);

CREATE INDEX IF NOT EXISTS idx_mdl_map_role_permission_permission
    ON public.mdl_map_role_permission (id_mst_permission);

-- Seed permissions (codes must match internal/entity/constant/permission/permission.go)
INSERT INTO public.mdl_mst_permission (code, resource, action, description)
SELECT v.code, v.resource, v.action, v.description
FROM (VALUES
    ('patient.read', 'patient', 'read', 'View patients'),
    ('patient.create', 'patient', 'create', 'Register patients'),
    ('patient.update', 'patient', 'update', 'Update patients'),
    ('visit.read', 'visit', 'read', 'View visits'),
    ('visit.create', 'visit', 'create', 'Create visits'),
    ('visit.update', 'visit', 'update', 'Update visits'),
    ('diagnosis.read', 'diagnosis', 'read', 'View diagnoses'),
    ('diagnosis.create', 'diagnosis', 'create', 'Create diagnoses'),
    ('diagnosis.update', 'diagnosis', 'update', 'Update diagnoses'),
    ('diagnosis.delete', 'diagnosis', 'delete', 'Delete diagnoses'),
    ('anamnesa.read', 'anamnesa', 'read', 'View anamnesa'),
    ('anamnesa.create', 'anamnesa', 'create', 'Create anamnesa'),
    ('anamnesa.update', 'anamnesa', 'update', 'Update anamnesa'),
    ('anamnesa.delete', 'anamnesa', 'delete', 'Delete anamnesa'),
    ('product.read', 'product', 'read', 'View products'),
    ('product.create', 'product', 'create', 'Create products'),
    ('product.update', 'product', 'update', 'Update products'),
    ('product.delete', 'product', 'delete', 'Delete products'),
    ('product.statistics', 'product', 'statistics', 'View product metrics and statistics'),
    ('journey.read', 'journey', 'read', 'View journey boards and points'),
    ('journey.create', 'journey', 'create', 'Create journey boards and points'),
    ('journey.update', 'journey', 'update', 'Update journey boards and points'),
    ('journey.delete', 'journey', 'delete', 'Delete journey boards and points'),
    ('recall.read', 'recall', 'read', 'View recalls'),
    ('recall.create', 'recall', 'create', 'Create recalls'),
    ('recall.update', 'recall', 'update', 'Update recalls'),
    ('recall.delete', 'recall', 'delete', 'Delete recalls'),
    ('odontogram.read', 'odontogram', 'read', 'View odontogram'),
    ('odontogram.create', 'odontogram', 'create', 'Create odontogram events'),
    ('odontogram.update', 'odontogram', 'update', 'Update odontogram'),
    ('odontogram.delete', 'odontogram', 'delete', 'Delete odontogram events'),
    ('reference.search', 'reference', 'search', 'Search ICD-10, doctors, and nurses')
) AS v(code, resource, action, description)
WHERE NOT EXISTS (
    SELECT 1 FROM public.mdl_mst_permission p WHERE p.code = v.code
);

-- Administrator: all permissions
INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
CROSS JOIN public.mdl_mst_permission p
WHERE r.name = 'administrator'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );

-- Clerk: patient, visit, recall, reference
INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code IN (
    'patient.read', 'patient.create', 'patient.update',
    'visit.read', 'visit.create', 'visit.update',
    'recall.read', 'recall.create', 'recall.update', 'recall.delete',
    'product.statistics',
    'reference.search'
)
WHERE r.name = 'clerk'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );

-- Doctor: patient.read, visit.*, diagnosis.*, anamnesa.*, odontogram.*, reference.search
INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code IN (
    'patient.read',
    'visit.read', 'visit.create', 'visit.update',
    'diagnosis.read', 'diagnosis.create', 'diagnosis.update', 'diagnosis.delete',
    'anamnesa.read', 'anamnesa.create', 'anamnesa.update', 'anamnesa.delete',
    'odontogram.read', 'odontogram.create', 'odontogram.update', 'odontogram.delete',
    'reference.search'
)
WHERE r.name = 'doctor'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );

-- Nurse: patient.read, visit.read, anamnesa.read, odontogram.*, recall.*, reference.search
INSERT INTO public.mdl_map_role_permission (id_mst_role, id_mst_permission)
SELECT r.id, p.id
FROM public.mdl_mst_role r
JOIN public.mdl_mst_permission p ON p.code IN (
    'patient.read',
    'visit.read',
    'anamnesa.read',
    'odontogram.read', 'odontogram.create', 'odontogram.update', 'odontogram.delete',
    'recall.read', 'recall.create', 'recall.update', 'recall.delete',
    'reference.search'
)
WHERE r.name = 'nurse'
  AND r.delete_time IS NULL
  AND p.delete_time IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.mdl_map_role_permission mrp
      WHERE mrp.id_mst_role = r.id
        AND mrp.id_mst_permission = p.id
  );
