# Staff compensation (payday)

Domain under `internal/{entity,http,usecase,repo}/compensation`. Specs: workspace `docs/prd|trd/staff-compensation.md`. Manual API: `docs/api/compensation-period.postman_collection.json`, `test_compensation_period_staff.sh`, `test_commission_item_patch.sh` (`MEDIANNE_TOKEN`).

## Routes (`/v1/compensation-period`, authed + RBAC)
| Method | Path | Perm |
| --- | --- | --- |
| POST | `/` | assign |
| GET | `/` `?status=&limit=&offset=` | read |
| GET | `/{periodId}` | read |
| GET | `/{periodId}/staffs` | read |
| GET | `/{periodId}/staff/{staffId}` | read |
| GET | `/{periodId}/staff/{staffId}/visits` | read |
| POST | `/{periodId}/staff/{staffId}/visits/generate` | assign |
| POST | `/{periodId}/draft` | assign |
| POST | `/{periodId}/finalize` | finalize |
| POST | `/{periodId}/reopen` | finalize |
| DELETE | `/{periodId}` | finalize |

Also (top-level under `/v1`, not nested under period):

| Method | Path | Perm |
| --- | --- | --- |
| PATCH | `/commission-items/{id}` | assign |

**Path trap:** list staff is **`/staffs`** (plural). Detail/visits use **`/staff/{staffId}`**. TRD sometimes says `/staff` for list — code wins.

Not wired yet (TRD only): wage-override, export. Batch commission update deferred.

## Permissions (`entity/constant/permission`)
`compensation.read` | `compensation.assign` | `compensation.finalize` | `compensation.manage`

## Lifecycle
Period status: `open` → `draft` → `finalized`. Reopen: finalized → draft (visits stay locked). Delete: open only (soft).
Dates: API `YYYY-MM-DD` (`compensationPeriodDateLayout`). Detection window: `[period_start, period_end+1day)` (end exclusive).

## Option A valuation
- Eligibility/sources only: procedure | diagnosis | anamnesa | journey | manual (`ContributionSource`).
- `revenue_base` = full visit product cart sum — **not** procedure-line filtered.
- Commission type `percent` | `flat`; amount resolved server-side on PATCH `/commission-items/{id}` (not on generate). Uses **stored** `revenue_base` (generate leaves 0 until a later card snapshots cart).

## Generate + list visits
- Generate seeds `mdl_trx_visit_commission` from detection; **skip existing** (idempotent).
- Seed defaults: type `flat`, flat/amount 0, `revenue_base` 0, **`approved_at` NULL**, sources JSON snapshot, `included_manually` from manual source.
- List visits: returns `id` (commission PK). If `approved_at` null → JSON nulls for `commission_type`, percent, flat, amount; still return sources/revenue_base/has_contributors.
- Schema: Atlas `schema/migrations/` (baseline includes compensation tables; `20260909120000` approved_at; archive: `schema/archive/medianne/20260901_add_compensation_tables.sql`).

## PATCH commission item
- `PATCH /v1/commission-items/{id}` (`compensation.assign`): updates one live row by id scoped to JWT institution via period.
- Writes only: `commission_type`, percent XOR flat, `note`, server `commission_amount`. Does **not** set `approved_at` or change `revenue_base` / `sources` / `included_manually`.
- Response: `{ updated_count: 1, commission_subtotal: 0 }` (real subtotal deferred).
- Errors: `COMMISSION_NOT_FOUND`, `INVALID_COMMISSION_TYPE`, `INVALID_COMMISSION_PERCENT`, `INVALID_COMMISSION_FLAT_AMOUNT`.

## Staff list semantics
- Union of period contributors (detection), not wage-only staff (v1).
- Wage/computed_wage stub **0**; override null until wage phase.
- `assignment_status`: unassigned | partial | complete from commissioned visit count vs detected visit count.
- Default list limit **50** (`defaultListLimit`).

## Tables / repo ownership
Compensation **writes**: `mdl_mst_staff_wage`, `mdl_trx_compensation_period`, `mdl_trx_visit_commission`, `mdl_map_visit_contributor`.
Visit lock cols (`compensation_period_id`, `compensation_locked_at`) live on `mdl_trx_patient_visit` → **patient** repo owns writes; compensation usecase may depend on a small lock interface, not update visit table from compensation repo.

## Packages
- Handler: `http/compensation/period_handler.go`
- UC: `usecase/compensation/period.go` (+ `period_test.go`)
- Repos: `repo/compensation/{period,commission,contributor,wage}_db.go`
- Models: `entity/model/compensation.go` (Xorm `sql.Null*` / API `null.v8` dual layer — `mem:conventions`)
- Visit contributors panel: `http|usecase/visit` + map table above

## Errors (common)
- `period_not_found` 400
- `staff_not_found` 404
- `INVALID_COMPENSATION_PERIOD_STATUS` 400
- `PERIOD_DATE_RANGE_OVERLAP` 400
- `ILLEGAL_PERIOD_TRANSITION` 400
- `COMMISSION_NOT_FOUND` 400
- `INVALID_COMMISSION_TYPE` 400
- `INVALID_COMMISSION_PERCENT` 400
- `INVALID_COMMISSION_FLAT_AMOUNT` 400
Envelope: `{ error_messages: [{ error_name, error_description }] }` (writer), success `{ data }`.

Map: `mem:core`. Style: `mem:conventions`.
