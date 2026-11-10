CREATE SCHEMA IF NOT EXISTS billing;

CREATE TABLE IF NOT EXISTS billing.invoices (
    id        uuid PRIMARY KEY,
    order_id  uuid        NOT NULL UNIQUE,
    amount    bigint      NOT NULL,
    currency  char(3)     NOT NULL,
    status    text        NOT NULL,
    issued_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS billing.inbox (
    event_id     text PRIMARY KEY,
    processed_at timestamptz NOT NULL
);
