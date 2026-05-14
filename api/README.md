# API

Stubbed REST API built with Go.

## Run

```bash
cd api
go run ./cmd/server
```

Or use Make:

```bash
cd api
make start
```

Server defaults to port `8080` and can be overridden with `PORT`.

## Routes

- `GET /v1/reporting-plans`

## Reporting Plans Query Route

`GET /v1/reporting-plans` runs the reporting plans query with these required filters:

- `ingestor_ids` (comma-separated UUIDs)
- `plan_id_types` (comma-separated strings)
- `plan_market_types` (comma-separated strings)

Example:

```bash
curl "http://localhost:8080/v1/reporting-plans?ingestor_ids=123e4567-e89b-12d3-a456-426614174000,223e4567-e89b-12d3-a456-426614174000&plan_id_types=EIN,HPID&plan_market_types=group,individual"
```

Set database connection details before calling this route:

- `PG_HOST` (required)
- `PG_USER` (required)
- `PG_PASSWORD` (required)
- `PG_PORT` (optional, defaults to `5432`)
- `PG_DATABASE` (optional, defaults to `postgres`)
- `PG_SSLMODE` (optional, defaults to `disable`)

## Make Targets

- `make build` builds `bin/api-server`
- `make start` runs the API with `go run`
- `make run` builds and runs the compiled binary
- `make test` runs all Go tests
- `make fmt` formats Go source files
- `make tidy` tidies module dependencies
- `make clean` removes build artifacts
