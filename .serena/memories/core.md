# Medianne backend (medilink)

Go API clinic DHR: patients, visits, clinical forms, inventory, journey board, odontogram, staff/RBAC, SatuSehat FHIR.
Module: `github.com/faisalhardin/medilink`. Path: workspace `backend/`; sibling UI `frontend/` (separate git).

## Source map
- `cmd/api` — HTTP server DI wire-up
- `cmd/job`, `cmd/cleanup` — batch/workers
- `internal/server` — Chi routes (`/v1/...`, `/ping`)
- `internal/http/<domain>` — handlers
- `internal/usecase/<domain>` — business logic
- `internal/repo/<domain>` — DB/cache/HTTP (Xorm/pgx, Redis/inmemory, SatuSehat)
- `internal/entity/{model,http,usecase,repo,constant}` — DTOs, interfaces, constants
- `internal/library` — middleware, errors, DB helpers, validation, idempotency
- `internal/config` — YAML + vault/env
- `schema/medianne/` — dated SQL migrations
- `files/etc/medilink/` — env YAMLs (dev yaml gitignored; use `*.example`)
- `docs/` — API/notes
- `tests/` — package tests (`make test` only runs here)

## Layer rule
Handlers → usecases → repos. Interfaces in `internal/entity/*`; impl in `internal/http|usecase|repo`. Wire in `cmd/api/main.go` + `internal/entity/http` modules.

## API surface
Prefix `/v1`. Auth `/v1/auth/*` (Google OAuth/goth, JWT refresh, sessions). Authed: institution, patient, visit (+ diagnosis/anamnesa/procedure), visit-detail, journey, odontogram, staff, recall, lookups (`/icd10|doctor|nurse|icd9cm/search`), admin/product. Health: `GET /ping`.

## Invariants
- JSON snake_case; Go `json` + `xorm` tags. DB cols often `id_mst_*`; soft-delete `delete_time`.
- Nullability: Xorm rows → `database/sql` (`sql.Null*`); JSON DTOs/payloads → `volatiletech/null/v8`. Details: `mem:conventions`.
- No SQL FK REFERENCES / ON DELETE on table rows (project rule).
- `time.Local = time.UTC` in API init.
- Institution-scoped multi-tenant; RBAC via `RequirePermission` + `entity/constant/permission`.
- SatuSehat FHIR R4 under `internal/entity/model/satusehat/`; queue/client in `repo/satusehat`.
- Deploy: Cloud Build → GCR → Cloud Run `asia-southeast1` (`cloudbuild.yaml`, `deploy.sh`).

## Related
Stack: `mem:tech_stack`. Commands: `mem:suggested_commands`. Style: `mem:conventions`. Done checks: `mem:task_completion`.