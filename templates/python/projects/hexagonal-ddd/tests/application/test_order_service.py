from collections.abc import Sequence
from datetime import UTC, datetime
from decimal import Decimal

import pytest

from sample_app.application.orders.commands import AddItem
from sample_app.application.orders.errors import OrderNotFoundError
from sample_app.application.orders.service import MAX_LIST_LIMIT, OrderService
from sample_app.domain.orders.events import OrderPlaced
from sample_app.domain.orders.order import Order
from sample_app.domain.shared.errors import BusinessRuleError
from sample_app.domain.shared.events import DomainEvent
from sample_app.domain.shared.identity import new_id

NOW = datetime(2026, 1, 1, tzinfo=UTC)


class FakeOrderRepository:
    def __init__(self) -> None:
        self.saved: dict[str, Order] = {}
        self.limits: list[int] = []

    def save(self, order: Order) -> None:
        self.saved[order.id] = order

    def find(self, order_id: str) -> Order | None:
        return self.saved.get(order_id)

    def list_recent(self, limit: int) -> list[Order]:
        self.limits.append(limit)
        return list(self.saved.values())[:limit]


class FakeEventPublisher:
    def __init__(self) -> None:
        self.published: list[DomainEvent] = []

    def publish(self, events: Sequence[DomainEvent]) -> None:
        self.published.extend(events)


@pytest.fixture
def repository() -> FakeOrderRepository:
    return FakeOrderRepository()


@pytest.fixture
def events() -> FakeEventPublisher:
    return FakeEventPublisher()


@pytest.fixture
def service(repository: FakeOrderRepository, events: FakeEventPublisher) -> OrderService:
    return OrderService(repository, events, lambda: NOW)


def test_placing_publishes_event(service: OrderService, events: FakeEventPublisher) -> None:
    order_id = service.create("Ana")
    service.add_item(order_id, AddItem("Caneta", 3, Decimal("2.5")))
    service.place(order_id)

    view = service.get(order_id)
    assert (view.status, view.total, view.currency) == ("Placed", Decimal("7.5"), "BRL")
    assert events.published == [OrderPlaced(order_id, 750, "BRL", NOW)]


def test_explicit_currency_is_kept(service: OrderService) -> None:
    order_id = service.create("Ana")
    service.add_item(order_id, AddItem("Pen", 1, Decimal("1"), "usd"))

    assert service.get(order_id).currency == "USD"


def test_rule_violation_does_not_save(service: OrderService, repository: FakeOrderRepository) -> None:
    order_id = service.create("Ana")
    with pytest.raises(BusinessRuleError):
        service.place(order_id)

    assert repository.saved[order_id].status == "Draft"


def test_unknown_order_is_not_found(service: OrderService) -> None:
    with pytest.raises(OrderNotFoundError):
        service.place(new_id())


def test_invalid_id_is_a_rule_violation(service: OrderService) -> None:
    with pytest.raises(BusinessRuleError):
        service.get("abc")


def test_list_limit_is_bounded(service: OrderService, repository: FakeOrderRepository) -> None:
    service.create("Ana")
    service.list_recent(0)
    service.list_recent(5000)
    service.list_recent(10)

    assert repository.limits == [MAX_LIST_LIMIT, MAX_LIST_LIMIT, 10]
