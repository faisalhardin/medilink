# Conventions

## Architecture
- Thin `http` handlers; logic in `usecase`; IO in `repo`.
- Interfaces + DTOs in `internal/entity/{usecase,repo,http,model}` by domain folder.
- New feature: entity → repo → usecase → http → `internal/server/routes.go` + wire `cmd/api/main.go`.
- Permissions: `internal/entity/constant/permission`; route with `RequirePermission(...)`.

## Naming / API
- Routes `/v1/...`.
- JSON snake_case; hide internals `json:"-"` (ids, audit times often omitted).
- Xorm quoted cols; FKs often `id_mst_institution`; soft delete `delete_time`.
- main import aliases: `foorepo`, `foouc`, `foohandler`.

## Models / nullability (Xorm vs JSON)
Dual-layer convention (see `procedure.go`, `compensation.go`, `anamnesa.go`):
- **Xorm / DB row structs:** nullable scalars use `database/sql` — `sql.NullInt64`, `sql.NullString`, `sql.NullTime`, `sql.NullFloat64`. Tag `xorm:"'col' null"`; entity fields usually `json:"-"`.
- **JSON request/response / JSONB payload structs:** nullable fields use `github.com/volatiletech/null/v8` — `null.Int64`, `null.String`, `null.Time`, `null.Float64` (not `sql.Null*`, not bare pointers for API nulls).
- **Soft delete:** `DeleteTime *time.Time` + xorm `deleted` (project-wide; do not switch soft-delete to sql.NullTime).
- **Optional JSONB on rows:** `json.RawMessage` (nil = SQL NULL).
- Map DB → API in usecase/DTO helpers (e.g. `TrxVisitProcedure.ToResponse()` copies `sql.Null*` → `null.*`).

## Errors / HTTP
- Structured errors `internal/library/common/commonerr` (`error_list`, codes).
- Writers `library/common/writer`. Middleware: `library/util/handler`, auth `library/middlewares/auth`, idempotency `library/idempotency` on sensitive creates.

## Data / SQL
- No table-row REFERENCES / ON DELETE FK (CLAUDE rule).
- Schema: new dated `schema/medianne/YYYYMMDD_*.sql`; avoid rewriting applied migrations.
- Minimal diffs; evidence from logs/API.

## Domain hotspots
- Visit owns nested diagnosis, anamnesa, procedure routes.
- Odontogram: event log + snapshot builder `usecase/odontogram`.
- SatuSehat FHIR shapes stay in `entity/model/satusehat`; queue separate from HTTP path.

## Style
- Go 1.19 idioms. Colocated `*_test.go` when present; Makefile suite is not full coverage.

Stack: `mem:tech_stack`. Map: `mem:core`.