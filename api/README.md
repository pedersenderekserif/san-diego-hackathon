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

- `GET /v1/reporting-plans/filters`
- `GET /v1/reporting-plans`

## Reporting Plan Filters Route

`GET /v1/reporting-plans/filters` returns dynamic values needed by the reporting plans query route:

- `ingestor_ids`
- `plan_id_types`
- `plan_market_types`

Example:

```bash
curl "http://localhost:8080/v1/reporting-plans/filters"
```

## Reporting Plans Query Route

`GET /v1/reporting-plans` runs the reporting plans query with these required filters:

- `ingestor_ids` (comma-separated UUIDs)
- `plan_id_types` (comma-separated strings)
- `plan_market_types` (comma-separated strings)

Optional filter:

- `eins` (comma-separated EINs; matches both dashed and non-dashed formats)

Example:

```bash
curl "http://localhost:8080/v1/reporting-plans?ingestor_ids=123e4567-e89b-12d3-a456-426614174000,223e4567-e89b-12d3-a456-426614174000&plan_id_types=EIN,HPID&plan_market_types=group,individual"
```

Filter by selected employer EIN while keeping required filters:

```bash
curl "http://localhost:8080/v1/reporting-plans?ingestor_ids=123e4567-e89b-12d3-a456-426614174000&plan_id_types=EIN&plan_market_types=group&eins=12-3456789"
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
