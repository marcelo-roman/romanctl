# Persistence

PostgreSQL with pgx, read from `DATABASE_URL`. `romanctl up sample-app` creates the `sample_app` database and writes the
variable to `.env`. Both services use the same database, each in its own schema, and neither reads the other's schema:

| Schema | Tables |
|---|---|
| `orders` | `orders`, `order_items` and `outbox` |
| `billing` | `invoices` (one per order) and `inbox` (the ids of events already processed) |

Each service creates its own schema at startup, under `pg_advisory_xact_lock`, so both can start at the same time.
`CREATE TABLE IF NOT EXISTS` does not evolve the schema: adopt migrations before you have data that matters.

The tests in `internal/services/*/adapters/postgres` and `internal/platform/postgres` run against the `DATABASE_URL`
database and are skipped without it.
