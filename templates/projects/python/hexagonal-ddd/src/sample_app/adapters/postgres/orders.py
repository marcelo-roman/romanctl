from collections import defaultdict
from typing import Any

from psycopg import Connection

from sample_app.adapters.postgres.database import Database
from sample_app.domain.orders.money import Money
from sample_app.domain.orders.order import Order
from sample_app.domain.orders.order_item import OrderItem
from sample_app.domain.orders.status import OrderStatus

UPSERT_ORDER = """
    INSERT INTO orders (id, customer_name, status, created_at) VALUES (%s, %s, %s, %s)
    ON CONFLICT (id) DO UPDATE SET customer_name = EXCLUDED.customer_name, status = EXCLUDED.status
"""
INSERT_ITEM = """
    INSERT INTO order_items (id, order_id, position, product_name, quantity, unit_price, currency)
    VALUES (%s, %s, %s, %s, %s, %s, %s)
"""
SELECT_ORDERS = "SELECT id, customer_name, status, created_at FROM orders"
SELECT_ITEMS = """
    SELECT order_id, id, product_name, quantity, unit_price, currency
    FROM order_items WHERE order_id = ANY(%s::uuid[]) ORDER BY position
"""


class PostgresOrderRepository:
    def __init__(self, database: Database) -> None:
        self._database = database

    def save(self, order: Order) -> None:
        with self._database.connection() as connection:
            connection.execute(UPSERT_ORDER, (order.id, order.customer_name, str(order.status), order.created_at))
            connection.execute("DELETE FROM order_items WHERE order_id = %s", (order.id,))
            rows = [
                (item.id, order.id, position, item.product_name, item.quantity, item.unit_price.cents,
                 item.unit_price.currency)
                for position, item in enumerate(order.items)
            ]
            if rows:
                connection.cursor().executemany(INSERT_ITEM, rows)

    def find(self, order_id: str) -> Order | None:
        orders = self._load(f"{SELECT_ORDERS} WHERE id = %s", (order_id,))
        return orders[0] if orders else None

    def list_recent(self, limit: int) -> list[Order]:
        return self._load(f"{SELECT_ORDERS} ORDER BY created_at DESC LIMIT %s", (limit,))

    def _load(self, query: str, params: tuple[Any, ...]) -> list[Order]:
        with self._database.connection() as connection:
            heads = connection.execute(query, params).fetchall()
            items = _items(connection, [str(head[0]) for head in heads])
        return [
            Order(str(order_id), customer_name, OrderStatus(status), created_at, items[str(order_id)])
            for order_id, customer_name, status, created_at in heads
        ]


def _items(connection: Connection, order_ids: list[str]) -> dict[str, list[OrderItem]]:
    items: dict[str, list[OrderItem]] = defaultdict(list)
    if not order_ids:
        return items
    for order_id, item_id, product_name, quantity, cents, currency in connection.execute(SELECT_ITEMS, (order_ids,)):
        items[str(order_id)].append(OrderItem(str(item_id), product_name, quantity, Money(cents, currency)))
    return items
