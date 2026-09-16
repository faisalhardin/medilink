# Project github.com/faisalhardin/medilink

One Paragraph of project description goes here

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## MakeFile

run all make commands with clean tests

```bash
make all build
```

build the application

```bash
make build
```

run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB container

```bash
make docker-down
```

## Database migrations (Atlas)

Schema is managed with **Atlas versioned migrations** (not ad-hoc `psql` of dated files).

| Path | Role |
|------|------|
| `schema/schema.sql` | Desired-state DDL |
| `schema/migrations/` | Versioned SQL + `atlas.sum` |
| `schema/archive/medianne/` | Frozen pre-Atlas history |
| `atlas.hcl` | Local / prod env config |

Full runbook: [`schema/README.md`](schema/README.md).

```bash
make install-atlas                    # once; pins CLI to ~/.local/bin
export PATH="$HOME/.local/bin:$PATH"

make docker-run                       # local Postgres
make migrate-status
make migrate-apply                    # empty local DB: baseline + follow-ups

# Change schema going forward:
#   1. edit schema/schema.sql
#   2. make migrate-diff NAME=add_foo
#   3. review schema/migrations/, then make migrate-apply
```

**Production** (Cloud SQL Proxy on `:5433`):

```bash
export DATABASE_URL='postgres://medianne-user:PASSWORD@127.0.0.1:5433/medianne?sslmode=disable'
# First cutover only (DB already at compensation tables baseline):
BASELINE=20260901120000 ./run-migrations-simple.sh
# Later:
./run-migrations-simple.sh
```

Never run `atlas schema apply` against prod — only `atlas migrate apply`.

live reload the application

```bash
make watch
```

run the test suite

```bash
make test
```

clean up binary from the last build

```bash
make clean
```
