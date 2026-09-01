# Tech stack

- **Language**: Go 1.19 (`go.mod`)
- **Module**: `github.com/faisalhardin/medilink`
- **HTTP**: chi/v5, chi/cors, gorilla/{schema,sessions}, markbates/goth (Google OAuth)
- **Auth**: cristalhq/jwt/v5; gothic cookie sessions
- **DB**: PostgreSQL (docker: `timescale/timescaledb:latest-pg15` :5432). Access: go-xorm/xorm + jackc/pgx/v5 + lib/pq. Soft delete via xorm `deleted`.
- **Cache**: Redis (go-redis/v8, redigo); wrapper `internal/repo/cache/inmemory`
- **Validation**: go-playground/validator/v10
- **Config**: gopkg.in/yaml.v2 + godotenv; vault nested under config
- **Logging**: rs/zerolog
- **Money/null**: shopspring/decimal, volatiletech/null/v8
- **Build**: Makefile → `go build -o main cmd/api/main.go`; optional air/realize reload
- **CI/deploy**: Dockerfile, docker-compose (psql), cloudbuild.yaml, GCloud Run
- **Vendor**: `vendor/` present
- **Tests**: sparse `*_test.go` under usecase/repo/pkg + `tests/handler_test.go`

Layout/invariants: `mem:core`.