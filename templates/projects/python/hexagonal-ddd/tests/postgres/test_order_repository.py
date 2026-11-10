import os
from datetime import UTC, datetime

from sample_app.adapters.postgres.database import Database
from sample_app.adapters.postgres.orders import PostgresOrderRepository
from sample_app.domain.orders.money import Money
from sample_app.domain.orders.order import Order
from sample_app.domain.orders.status import OrderStatus
from sample_app.domain.shared.identity import new_id


def test_order_round_trip(database: Database) -> None:
    repository = PostgresOrderRepository(database)
    order = Order.create("Ana", datetime.now(UTC).replace(microsecond=0))
    order.add_item("Caneta", 2, Money(350, "BRL"))
    order.add_item("Caderno", 1, Money(1200, "BRL"))
    repository.save(order)

    loaded = repository.find(order.id)

    assert loaded is not None
    assert loaded.customer_name == "Ana"
    assert loaded.created_at == order.created_at
    assert loaded.total == Money(1900, "BRL")
    assert [item.product_name for item in loaded.items] == ["Caneta", "Caderno"]
    assert repository.find(new_id()) is None


def test_saving_again_updates_status_and_items(database: Database) -> None:
    repository = PostgresOrderRepository(database)
    order = Order.create("Bia", datetime.now(UTC))
    repository.save(order)
    order.add_item("Lápis", 3, Money(100, "BRL"))
    order.place(datetime.now(UTC))
    repository.save(order)

    loaded = repository.find(order.id)

    assert loaded is not None
    assert loaded.status == OrderStatus.PLACED
    assert loaded.total == Money(300, "BRL")
    assert order.id in [recent.id for recent in repository.list_recent(200)]


def test_schema_creation_is_idempotent(database: Database) -> None:
    again = Database(os.environ["DATABASE_URL"])
    again.open()
    again.close()
