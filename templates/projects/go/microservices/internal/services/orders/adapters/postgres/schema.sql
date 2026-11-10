CREATE SCHEMA IF NOT EXISTS orders;

CREATE TABLE IF NOT EXISTS orders.orders (
    id            uuid PRIMARY KEY,
    customer_name text        NOT NULL,
    status        text        NOT NULL,
    created_at    timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS orders.order_items (
    id           uuid PRIMARY KEY,
    order_id     uuid    NOT NULL REFERENCES orders.orders (id) ON DELETE CASCADE,
    position     integer NOT NULL,
    product_name text    NOT NULL,
    quantity     integer NOT NULL,
    unit_price   bigint  NOT NULL,
    currency     char(3) NOT NULL
);

CREATE INDEX IF NOT EXISTS order_items_order_id ON orders.order_items (order_id);

CREATE TABLE IF NOT EXISTS orders.outbox (
    id          uuid PRIMARY KEY,
    type        text        NOT NULL,
    payload     jsonb       NOT NULL,
    occurred_at timestamptz NOT NULL,
    sent_at     timestamptz
);

CREATE INDEX IF NOT EXISTS outbox_pending ON orders.outbox (occurred_at) WHERE sent_at IS NULL;

-- romanctl:tables
