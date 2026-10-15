# Staff compensation (worksheet + payday)

Domain under `internal/{entity,http,usecase,repo}/compensation`. Specs: workspace `docs/prd|trd/staff-compensation.md`.

## Product split
- **Worksheet** (`mdl_trx_worksheet`): single-staff wrap for visit commissions. Create default **`open` + `idle`**. Primary place to generate / edit / finalize commissions.
- **Payday period** (`mdl_trx_compensation_period`): rolls up worksheets via `worksheet.compensation_period_id`. Does **not** generate commissions, lock visits, or generate wage snapshots.

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
- Handlers: `http/compensation/{period,worksheet,visit_commission,wage,wage_snapshot}_handler.go`
- UC: `usecase/compensation/{period,worksheet,visit_commission,finalize,wage,wage_snapshot}.go`
- Repos: `repo/compensation/{period,worksheet,commission,contributor,wage,wage_snapshot}_db.go`; visit lock in `repo/patient`
- Utils: `library/util/common/time.go` (`DateOnly`, `PeriodsOverlap`)
- Models: `entity/model/compensation.go`

## Wages (`mdl_mst_staff_wage`)
- `GET|PUT /v1/compensation/wages`, `DELETE /v1/compensation/wages/{wageId}`. Perm `compensation.manage`.
- Cadence `monthly` | `weekly` | `daily`.
- One active wage per staff+institution. PUT inserts a new active row and closes the previous one the day before the new `effective_from`. Reusing that start date → `WAGE_EFFECTIVE_RANGE_OVERLAP`. `effective_to` null = open-ended.
- GET is active rows only. `created_at` on `StaffWageResponse` is RFC3339 (`ToResponse`). Upsert sets `CreateTime` before insert; the xorm `created` tag does not fill the PUT body.
- Errors: `INVALID_WAGE_CADENCE`, `WAGE_EFFECTIVE_RANGE_INVALID`, `WAGE_EFFECTIVE_RANGE_OVERLAP`, `WAGE_MULTIPLE_ACTIVE`, `WAGE_NOT_FOUND`.
- `WageDB.ListLive` is every non-deleted wage. `ListActive` is the wage list API. `ListLiveByStaff` is the overlap check on PUT.

## Wage snapshots (`mdl_trx_wage_period_snapshot`)
- Perm `compensation.manage`. `POST /v1/compensation/staff-wages/generate`, `GET /`, `PATCH /{id}`, `DELETE /{id}`.
- Generate loads `ListLive` and keeps wages whose effective range overlaps the payday dates (`PeriodsOverlap` / `wageRangeEnd`). `staff_ids` empty → those staff. A named staff with none → 400 `WAGE_CONTRACT_NOT_FOUND`.
- Two contracts overlapping the period: later `effective_from`, then higher id, is copied for the whole period. Amount and cadence are copied; day counts and `total_wage` start null.
- PATCH: both day counts set → `total_wage = wage_amount * staff_working_days / unit days` (trunc toward zero; mandatory must be > 0). A week is 7 days, a day is 1, a month is the mandatory days. Otherwise a set `final_wage` is copied. A missing input stays null. Negative days or final wage → `WAGE_SNAPSHOT_DAYS_INVALID` / `WAGE_SNAPSHOT_FINAL_WAGE_INVALID`.
- PATCH and DELETE move the payday `total_wage` by the snapshot difference (null is 0) and set `total_payout` to that wage plus `total_commission`. Draft and finalize keep `total_wage`. They do not generate snapshots.

## Visit lock enforcement
- `compensation_locked_at` is written only by worksheet finalize (`patient.LockVisits`). Payday draft/finalize do not set it, so draft periods do not lock visits.
- Helper: `usecase/visit.RejectIfLocked`. Null timestamp → allow. Set timestamp → HTTP 403 `VISIT_COMPENSATION_LOCKED` / `visit compensation is locked` for every role, including administrator. No audit log. Envelope stays `error_messages`.
- Gated writes: anamnesa upsert; diagnosis save and delete; procedure save and delete; visit-product insert, upsert, and update; touchpoint upsert; `PATCH /v1/visit/{id}`; `PATCH /v1/visit/archive`; contributor add and delete.
- Reads stay open. Do not enforce inside `authorizeVisit` or `ValidatePatientVisitExist`. `GetPatientVisits` SELECT omits `compensation_locked_at`; those writers re-read with `GetPatientVisitsByID`.

## Common errors
`WORKSHEET_NOT_FOUND`, `WORKSHEET_DATE_RANGE_OVERLAP`, `WORKSHEET_NOT_EDITABLE`, `WORKSHEET_GENERATE_PENDING`, `WORKSHEET_LINKED_TO_PAYDAY`, `PAYDAY_PERIOD_NOT_FOUND`, `PAYDAY_PERIOD_FINALIZED`, `COMMISSION_WORKSHEET_LINKED_TO_PAYDAY`, `COMMISSION_WORKSHEET_FINALIZED`, `FORBIDDEN`, `VISIT_COMPENSATION_LOCKED`, period errors.

Map: `mem:core`. Style: `mem:conventions`. Maintenance: `mem:memory_maintenance`.
