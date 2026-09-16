-- Enforce global uniqueness of active staff emails for deterministic SSO login
-- (soft-deleted staff can reuse old emails after reactivation policy decisions).

CREATE UNIQUE INDEX IF NOT EXISTS uidx_mdl_mst_staff_active_email
    ON public.mdl_mst_staff (LOWER(email))
    WHERE delete_time IS NULL;
