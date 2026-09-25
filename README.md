# Basket Order Module — Go

Go implementation of the supplied Java module, with all **36 active REST routes**, YAML configuration, GORM, and MySQL (default), PostgreSQL, SQL Server, and SQLite drivers. The source is organized by business module under `internal/app`; DTOs and database mappings are under `internal/model`.

Start with the [complete API reference](docs/API_REFERENCE.md), [API inventory](docs/API.md), [OpenAPI](docs/openapi.json), and [behavioral differences and integration notes](docs/PARITY.md). [Java source coverage](docs/SOURCE_COVERAGE.md) accounts for all 138 Java source files. Route/source inventories are traceability checks, not a claim of proven production equivalence.

## Run

Go 1.26 or newer is required by the pinned dependencies. SQLite uses CGO; install a C compiler if running SQLite or the default tests.

```sh
cd go
go mod download
export BASKET_DB_DSN='user:password@tcp(127.0.0.1:3306)/basket?charset=utf8mb4&parseTime=true&loc=Asia%2FKolkata'
export BASKET_OIDC_ISSUER='https://identity.example/realms/trading'
export BASKET_OIDC_AUDIENCE='basket-client'
export BASKET_ADMIN_TOKEN='your-admin-authorization-value'
export BASKET_CACHE_CLUSTER='your-existing-cluster'
# Set the upstream URLs and Firebase settings referenced in config.yaml.
go run ./cmd/basket -config config.yaml -check-config
go run ./cmd/basket -config config.yaml
```

The default HTTP port is **9009**, matching Java. Configuration values can be literal YAML scalars or `${ENVIRONMENT_VARIABLE}` references. Expansion happens after YAML parsing so punctuation/newlines in secrets cannot inject configuration. Unknown YAML keys fail validation. Secrets from the Java properties file are not copied into Go configuration.

For a local run without external services:

```sh
go run ./cmd/basket -config testdata/config.local.yaml
```

The local configuration uses SQLite, a file-backed sample contract cache, an explicit `LOCALUSER` identity, and disables notifications and scheduled jobs. Create/list/update/delete, research reads, thematic previews, and holdings can run locally. Order placement and remote margin calls require configured upstream services; local mode does not fake their success.

## Database configuration

Only `database.driver` and `database.dsn` change between engines:

| Driver | Example DSN |
|---|---|
| `mysql` (default) | `user:password@tcp(127.0.0.1:3306)/basket?charset=utf8mb4&parseTime=true&loc=Asia%2FKolkata` |
| `postgres` / `postgresql` | `host=127.0.0.1 port=5432 user=basket password=secret dbname=basket sslmode=require` |
| `mssql` / `sqlserver` | `sqlserver://user:password@127.0.0.1:1433?database=basket&encrypt=true` |
| `sqlite` | `file:basket.db?_foreign_keys=on&_busy_timeout=5000` |

Use the driver-specific DSN to configure TLS and connection options. Pool settings live alongside the DSN. Table/column names retain the source's SQL names, normalized to lowercase for cross-engine use. The SQL Server implementation uses update-lock table hints when enforcing basket limits; MySQL/PostgreSQL use `FOR UPDATE`.

`auto_migrate` defaults to `false` for existing installations. For a new/test database:

```sh
go run ./cmd/basket -config config.yaml -migrate
```

Migration includes JPA entities and JDBC-only thematic tables/columns. It creates/updates schema; it does **not** copy records from MySQL to another engine. The supplied project contains no canonical SQL migrations, so compare the inferred schema to your existing database before adopting it. GORM driver references: [connections](https://gorm.io/docs/connecting_to_the_database.html), [transactions](https://gorm.io/docs/transactions.html).

## Module switches

Each switch in `modules` accepts `true` or `false`. Disabled API modules do not register their routes (HTTP 404).

| Switch | Scope |
|---|---|
| `basket` | Personal basket CRUD, scrip changes, execute/reset |
| `admin` | Vendor validation, research-basket distribution, expired-basket deletion |
| `research` | Research calls, status lists, sector reports |
| `thematic` | Catalog/details, invest/v1 invest, review, rebalance, views, legacy holdings |
| `holdings` | Both `/thematic/holdings/get` variants |
| `margin` | Span and NEST basket margin |
| `cache` | Expired-scrip maintenance API and its cleanup job |
| `notifications` | Push delivery and notification-record creation |
| `scheduler` | Configurable daily cleanup schedule |
| `token` | Token diagnostics and logout response |

The `cache` module switch controls the maintenance feature; trading modules still need their contract/session cache provider. `scheduler.delete_scrips`, `scheduler.delete_baskets`, and `scheduler.startup_cleanup` independently control cleanup behavior. The default cleanup schedule is 01:00 in the configured timezone.

## Integration settings

- **Authentication:** OIDC discovery/JWKS with optional required user-info lookup (enabled by default, matching Java), or HS256 JWT verification; issuer, audience, expiration, and the configured user claim are checked. User IDs are uppercased as in Java. Admin creation uses the configured raw `Authorization` value plus the request's vendor `apiKey`. The source's public expired-admin-basket route is listed explicitly in `auth.public_paths`; remove that entry to require JWT authentication.
- **Cache:** native Hazelcast supports the supplied Java `Serializable` contract/customer classes and Java-compatible device-list writes, as well as JSON values. `http` provides an optional JSON bridge boundary; `file` is for local fixtures. See PARITY.md for protocol limits.
- **Orders:** set `upstream.order_url` to the full bulk-order endpoint. The incoming bearer token is forwarded; thematic orders include `TBK:<basketId>` and become AMO outside configured trading days/hours. No automatic order retry occurs.
- **Margins/order book:** set full upstream URLs; NEST session keys and Tomcat routing values come from the customer cache. `legacy_nest_first_only` defaults to Java's behavior; disabling it includes all submitted legs.
- **Notifications:** FCM v1 uses a service-account file and project ID; `legacy` supports the older API-key payload. Admin `/temp` persists notification records without push delivery. Background campaigns use a bounded queue and finish during graceful shutdown.
- **Logging:** stdout access logs, optional SQL audit tables, and optional ClickHouse HTTP insertion. `logging.database_config` can point to a separate GORM database. SQL audit tables retain Java's hourly/daily names and must already exist. Log bodies are bounded and credentials are redacted. TLS validation is enabled; a custom CA file or explicit compatibility override is configurable.

## Verification

Run the complete HTTP smoke/dry test and generate an endpoint report:

```sh
make smoke
```

[Latest smoke report](docs/SMOKE_REPORT.md) and [response details](docs/smoke-results.json) cover all 36 routes, missing credentials, and disabled modules. The harness uses loopback HTTP, disposable SQLite, signed test JWTs, and local broker/margin/notification mocks. Outbound application HTTP is restricted to the mock server. No real orders or notifications are sent.

```sh
go test -race ./...
go vet ./...
go build ./cmd/basket
python3 tools/audit_parity.py
```

Tests cover basket lifecycle/ownership/limits, execution and reset, module switches, authentication failures, scalar/date handling, expiry cleanup, admin temporary creation, research visibility/filtering, thematic previews/execution/market hours, holdings recommendation differences, NEST margins, order-book updates, Java serialization fixtures, and config validation.

Disposable database integration tests:

```sh
docker compose -p basket-go-test -f testdata/compose.yaml up -d --wait mysql postgres
BASKET_TEST_MYSQL_DSN='root:basket-test-only@tcp(127.0.0.1:13316)/basket_test?parseTime=true' \
BASKET_TEST_POSTGRES_DSN='host=127.0.0.1 port=15442 user=postgres password=basket-test-only dbname=basket_test sslmode=disable' \
go test ./internal/app -run TestDatabaseIntegration -v
# SQL Server is available with --profile mssql; create basket_test first,
# then set BASKET_TEST_MSSQL_DSN to the disposable database.
docker compose -p basket-go-test -f testdata/compose.yaml --profile mssql down -v
```

The integration suite mutates its explicitly supplied test databases. It never uses `BASKET_DB_DSN`. Java/Hazelcast interoperability fixtures are generated from the supplied runner JAR; see `internal/javaser/testdata`. Java is needed only to regenerate or run those interoperability fixtures, not to run the Go service.
