# Task completion checks

From `backend/` after meaningful changes:

1. **Compile**: `go build -o /dev/null ./cmd/api/` or `make build`
2. **Tests**: `go test ./tests -v`; touched pkgs `go test ./internal/usecase/<pkg>/... ./internal/repo/<pkg>/...` or `go test ./...`
3. **Vet** (optional): `go vet ./...`
4. **Routes/DI**: new endpoint registered in `internal/server/routes.go` and wired in `cmd/api/main.go`
5. **Schema**: new SQL in `schema/medianne/`; migrations still apply (`./check_migrations.sh` when relevant)
6. **Contract**: JSON tags match frontend models when cross-stack
7. **Secrets**: never commit `*.development.yaml`, vault, `.env`

No special formatter gate beyond gofmt — `gofmt -w` on edited `.go` files.

Commands: `mem:suggested_commands`.