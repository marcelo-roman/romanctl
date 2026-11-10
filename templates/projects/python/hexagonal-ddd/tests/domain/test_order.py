from datetime import UTC, datetime
from decimal import Decimal

import pytest

from sample_app.domain.orders.events import OrderPlaced
from sample_app.domain.orders.money import Money
from sample_app.domain.orders.order import Order
from sample_app.domain.orders.status import OrderStatus
from sample_app.domain.shared.errors import BusinessRuleError

NOW = datetime(2026, 1, 1, 12, 0, tzinfo=UTC)


def brl(cents: int) -> Money:
    return Money(cents, "brl")


def draft() -> Order:
    return Order.create("Ana", NOW)


def test_create_requires_customer() -> None:
    with pytest.raises(BusinessRuleError):
        Order.create("  ", NOW)


def test_new_order_is_an_empty_draft() -> None:
    order = Order.create("  Ana ", NOW)

    assert order.customer_name == "Ana"
    assert order.status == OrderStatus.DRAFT
    assert order.created_at == NOW
    assert order.total == Money(0, "BRL")


def test_total_sums_items() -> None:
    order = draft()
    order.add_item("Caneta", 2, brl(350))
    order.add_item("Caderno", 1, brl(1200))

    assert order.total == Money(1900, "BRL")
    assert order.items[0].subtotal == Money(700, "BRL")


def test_item_requires_product_and_positive_quantity() -> None:
    order = draft()
    with pytest.raises(BusinessRuleError):
        order.add_item(" ", 1, brl(100))
    with pytest.raises(BusinessRuleError):
        order.add_item("Caneta", 0, brl(100))


def test_items_must_share_currency() -> None:
    order = draft()
    order.add_item("Caneta", 1, brl(100))
    with pytest.raises(BusinessRuleError):
        order.add_item("Pen", 1, Money(100, "USD"))


def test_place_raises_event_once() -> None:
    order = draft()
    with pytest.raises(BusinessRuleError):
        order.place(NOW)
    order.add_item("Caneta", 1, brl(100))
    order.place(NOW)

    assert order.status == OrderStatus.PLACED
    assert order.pull_events() == [OrderPlaced(order.id, 100, "BRL", NOW)]
    assert order.pull_events() == []


def test_placed_order_cannot_change() -> None:
    order = draft()
    order.add_item("Caneta", 1, brl(100))
    order.place(NOW)
    with pytest.raises(BusinessRuleError):
        order.add_item("Lápis", 1, brl(100))
    with pytest.raises(BusinessRuleError):
        order.place(NOW)


def test_money_rules() -> None:
    with pytest.raises(BusinessRuleError):
        Money(-1, "BRL")
    with pytest.raises(BusinessRuleError):
        Money(100, "REAL")
    with pytest.raises(BusinessRuleError):
        Money(100, "BRL").add(Money(100, "USD"))
    assert Money.of(Decimal("3.5")) == Money(350, "BRL")
    assert Money.of(Decimal("0.125"), "usd") == Money(13, "USD")
    assert Money(350, "BRL").amount == Decimal("3.5")
