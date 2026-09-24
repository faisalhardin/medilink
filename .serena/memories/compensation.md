# Staff compensation (worksheet + payday)

Domain under `internal/{entity,http,usecase,repo}/compensation`. Specs: workspace `docs/prd|trd/staff-compensation.md`.

## Product split
- **Worksheet** (`mdl_trx_worksheet`): single-staff wrap for visit commissions. Create default **`open` + `idle`**. Primary place to generate / edit / finalize commissions.
- **Payday period** (`mdl_trx_compensation_period`): rolls up worksheets via `worksheet.compensation_period_id`. Does **not** generate commissions or lock visits. Wages still Phase-1 stub (0).

## Worksheet lifecycle
- `status`: `pending` | `open` | `finalized`
- `generate_status`: `idle` | `running` | `succeeded` | `failed`
- Create → `open` + `idle`
- Generate (async insert-if-missing, Option A `revenue_base` = full visit product cart via `SumRevenueByVisitIDs`):
  - `POST /v1/worksheet/{id}/generate` (assign) → always starts worker
  - `POST /v1/visit-commissions/generate { worksheet_id, include_generate }` → worker only when `include_generate: true` (202); `false` → 200 current status
  - marks `pending`/`running` → finish `open` + `succeeded`|`failed`
- Finalize worksheet → `finalized` + lock visits (`worksheet_id` + `compensation_locked_at` on visit)
- `visit_count` = distinct **all live** commission visits; `total_commission` = **approved-only** sum

## Routes

### `/v1/worksheet` (compensation RBAC)
| Method | Path | Perm | Notes |
| --- | --- | --- | --- |
| POST | `/` | assign | create open/idle |
| GET | `/` `?staff_id=&status=&compensation_period_uuid=&limit=&cursor=` | read | `{ worksheets, next_cursor }` no total |
| GET | `/{id}` | read | |
| GET | `/{id}/commissions` | read | |
| POST | `/{id}/generate` | assign | async |
| PATCH | `/{id}` | assign | open only; `compensation_period_uuid` *null.String: nil=omit, invalid/empty=detach, valid=attach |
| DELETE | `/{id}` | assign | not finalized, not payday-linked; TX cascade SoftDeleteByWorksheet then worksheet |
| POST | `/{id}/finalize` | finalize | |

### `/v1/visit-commissions` (authed; UC admin / compensation.* OR self)
| Method | Path | Notes |
| --- | --- | --- |
| POST | `/generate` | `{ worksheet_id, include_generate }` |
| GET | `/` | `?staff_id=&start=&end=`; response uses public `worksheet_uuid` (not internal id) |
| PATCH | `/{id}` | open worksheet only; **blocked if payday-linked**; refresh Option A revenue_base then resolve amount; persists revenue_base |
| DELETE | `/{id}` | soft archive; blocked if payday-linked or finalized |

### `/v1/compensation-period` (payday rollup; no commission generate)
Standard CRUD + draft/finalize/reopen. List staff via worksheets rollup. Path trap: list staff is **`/staffs`**.

## Option A valuation
- Sources = eligibility only.
- `revenue_base` = SUM full visit cart (`mdl_trx_visit_product`) at generate seed and again at PATCH assignment.
- Type `percent`|`flat`; amount server-side on PATCH.

## Packages
- Handlers: `http/compensation/{period,worksheet,visit_commission}_handler.go`
- UC: `usecase/compensation/{period,worksheet,visit_commission,finalize}.go`
- Repos: `repo/compensation/{period,worksheet,commission,contributor,wage}_db.go`; visit lock in `repo/patient`
- Utils: `library/util/common/time.go` (`DateOnly`, `PeriodsOverlap`)
- Models: `entity/model/compensation.go`

## Common errors
`WORKSHEET_NOT_FOUND`, `WORKSHEET_DATE_RANGE_OVERLAP`, `WORKSHEET_NOT_EDITABLE`, `WORKSHEET_GENERATE_PENDING`, `WORKSHEET_LINKED_TO_PAYDAY`, `PAYDAY_PERIOD_NOT_FOUND`, `PAYDAY_PERIOD_FINALIZED`, `COMMISSION_WORKSHEET_LINKED_TO_PAYDAY`, `COMMISSION_WORKSHEET_FINALIZED`, `FORBIDDEN`, period errors.

Map: `mem:core`. Style: `mem:conventions`. Maintenance: `mem:memory_maintenance`.
