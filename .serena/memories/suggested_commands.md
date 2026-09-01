# Suggested commands (from `backend/`)

## App
- `make run` / `go run cmd/api/main.go` — API
- `make build` — binary `./main`
- `make watch` — air live reload (prompts install if missing)
- `make app` — `.env-dist`→`.env`, realize start `medilink` (ISLOCAL=1)
- `go run cmd/job/main.go` / `go run cmd/cleanup/main.go` — workers

## Deps
- `go mod tidy` from backend root (module medilink, not workspace root)

## DB / infra (Darwin)
- `make docker-run` / `make docker-down` — Timescale/Postgres compose
- Redis: `./setup-redis.sh` if needed
- Migrations: `./run-migrations.sh`, `./run-migrations-simple.sh`, `./manual-migrations.sh`, `./scripts/migrate.sh`, `./check_migrations.sh` — SQL in `schema/medianne/`
- Cloud SQL: `./start-cloud-sql-proxy.sh`

## Config
- Copy `files/etc/medilink/medilink.development.yaml.example` → `medilink.development.yaml` (gitignored)
- Default API `127.0.0.1:8080`; frontend callback often `http://127.0.0.1:5173`

## Test / clean
- `make test` → `go test ./tests -v` only
- Broader: `go test ./...`
- `make clean` — remove `main`

## Deploy
- `./deploy.sh` — build/push Cloud Run (needs gcloud)

Darwin: prefer `docker compose`; Makefile falls back to `docker-compose`.

Done gates: `mem:task_completion`.