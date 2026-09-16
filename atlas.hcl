// Atlas versioned migrations for Medianne.
// Desired state: schema/schema.sql
// Versioned SQL: schema/migrations/
//
// Never run `atlas schema apply` against prod — only `atlas migrate apply`.

variable "database_url" {
  type    = string
  default = getenv("DATABASE_URL")
}

locals {
  // TimescaleDB installs these schemas on the compose image; ignore them.
  timescale_exclude = [
    "_timescaledb_cache",
    "_timescaledb_catalog",
    "_timescaledb_config",
    "_timescaledb_internal",
    "timescaledb_experimental",
    "timescaledb_information",
  ]
}

env "local" {
  url = var.database_url != "" ? var.database_url : "postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable"
  // Dev DB for migrate diff / lint (plain Postgres 15; extensions via docker block if needed).
  dev = "docker://postgres/15/atlas_dev?search_path=public"
  src = "file://schema/schema.sql"
  exclude = local.timescale_exclude
  migration {
    dir = "file://schema/migrations"
  }
}

env "prod" {
  // Cloud SQL Proxy: ./start-cloud-sql-proxy.sh → 127.0.0.1:5433
  // Example: postgres://medianne-user:${DB_PASSWORD}@127.0.0.1:5433/medianne?sslmode=disable
  url = var.database_url
  dev = "docker://postgres/15/atlas_dev?search_path=public"
  src = "file://schema/schema.sql"
  exclude = local.timescale_exclude
  migration {
    dir = "file://schema/migrations"
  }
}
