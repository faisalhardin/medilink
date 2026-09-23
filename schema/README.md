# Database schema (Atlas)

Medianne uses **Atlas versioned migrations**. Desired DDL lives in `schema.sql`; apply history lives in `migrations/`. Historical dated SQL is frozen under `archive/medianne/` (do not add new files there).

**Never** run `atlas schema apply` against production. Only `atlas migrate apply`.

## Layout

```
schema/
  schema.sql                 # desired-state DDL (source for migrate diff)
  migrations/                # versioned SQL + atlas.sum
  archive/medianne/          # frozen pre-Atlas dated SQL
  README.md                  # this runbook
```

Config: [`../atlas.hcl`](../atlas.hcl). Makefile targets from `backend/`.

## Install Atlas (once)

```bash
cd backend
make install-atlas          # pins ATLAS_VERSION (default v0.35.0) into ~/.local/bin
export PATH="$HOME/.local/bin:$PATH"
atlas version
```

## Day-to-day workflow

1. Edit DDL in `schema/schema.sql` (not a new dated file under archive).
2. Generate a migration (needs Docker for the Atlas dev DB):

   ```bash
   make migrate-diff NAME=add_foo
   ```

3. Review `schema/migrations/*_add_foo.sql`. For `CREATE INDEX CONCURRENTLY`, put this at the top **followed by two blank lines** (required for Atlas to parse the directive):

   ```sql
   -- atlas:txmode none


   CREATE INDEX CONCURRENTLY ...
   ```

4. After any hand-edit of migration SQL:

   ```bash
   make migrate-hash
   ```

5. Apply locally:

   ```bash
   make migrate-status
   make migrate-apply
   ```

Do **not** rewrite migration files that are already recorded in `atlas.sum` and applied to shared environments. Add a new version instead.

## Local environments

### Empty database

```bash
make docker-run             # Timescale/Postgres on :5432
make migrate-apply          # applies baseline + all follow-ups
```

### Database already fully ahead of Atlas (all follow-ups already applied manually)

Record the latest version without re-running SQL:

```bash
make migrate-status
make migrate-set VERSION=20260909130000
```

### Database at production lag (through compensation tables only)

Same as prod first cutover: mark baseline, then apply pending:

```bash
make migrate-apply BASELINE=20260901120000
```

## Production cutover

Production is at `20260901_add_compensation_tables` (Atlas baseline `20260901120000`). Pending versions after baseline:

| Version | Purpose |
|---------|---------|
| `20260902120000` | `idx_trx_visit_product_visit` (CONCURRENTLY) |
| `20260907120000` | `uidx_mst_staff_uuid` |
| `20260907130000` | `idx_trx_patient_visit_institution_create_time` |
| `20260908120000` | procedure `doctor_id`/`nurse_id` → UUID (+ index) |
| `20260909120000` | `approved_at` on visit commission |
| `20260909130000` | compensation permission seed |

Steps:

1. Start Cloud SQL Proxy: `./start-cloud-sql-proxy.sh` (TCP `127.0.0.1:5433`).
2. Set URL:

   ```bash
   export DATABASE_URL='postgres://medianne-user:PASSWORD@127.0.0.1:5433/medianne?sslmode=disable'
   ```

3. Dry status:

   ```bash
   make migrate-status ATLAS_ENV=prod
   ```

4. **First cutover only** — mark baseline applied, then run pending:

   ```bash
   BASELINE=20260901120000 ./run-migrations-simple.sh
   # or: make migrate-apply ATLAS_ENV=prod BASELINE=20260901120000
   ```

5. Later deploys (no baseline flag):

   ```bash
   ./run-migrations-simple.sh
   ```

6. Confirm: `make migrate-status ATLAS_ENV=prod`

Before `20260908120000`, if `doctor_id`/`nurse_id` are still varchar, fix invalid UUID strings (see comments in that migration).

## Optional: polish `schema.sql` from a live DB

The checked-in `schema.sql` is an assembled starting point (IF NOT EXISTS style). For a clean inspect dump after migrations are applied:

```bash
./scripts/atlas_inspect_schema.sh
# writes schema/schema.sql from DATABASE_URL (default local)
make migrate-hash   # only if you also changed migrations/
```

## Helpers

| Command | Role |
|---------|------|
| `make install-atlas` | Pin/install Atlas CLI |
| `make migrate-diff NAME=…` | Generate migration from `schema.sql` |
| `make migrate-hash` | Refresh `atlas.sum` |
| `make migrate-status` | Show applied / pending |
| `make migrate-apply` | Apply pending (`BASELINE=` optional) |
| `make migrate-set VERSION=…` | Record version without SQL |
| `make migrate-lint` | Lint latest migration |
| `./scripts/migrate.sh` | Local apply wrapper |
| `./run-migrations-simple.sh` | Prod apply (proxy + `DATABASE_URL`) |
| `./scripts/build_atlas_baseline.sh` | Rebuild baseline concat from archive (rare) |
| `./scripts/atlas_inspect_schema.sh` | Regenerate `schema.sql` from a DB |

Deprecated (exit 1 with pointer here): `./run-migrations.sh`, `./manual-migrations.sh`.
