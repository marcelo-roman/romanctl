CREATE TABLE IF NOT EXISTS orders (
    id            uuid PRIMARY KEY,
    customer_name text        NOT NULL,
    status        text        NOT NULL,
    created_at    timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS order_items (
    id           uuid PRIMARY KEY,
    order_id     uuid    NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    position     integer NOT NULL,
    product_name text    NOT NULL,
    quantity     integer NOT NULL,
    unit_price   bigint  NOT NULL,
    currency     char(3) NOT NULL
);

CREATE INDEX IF NOT EXISTS order_items_order_id ON order_items (order_id);

-- romanctl:tables
