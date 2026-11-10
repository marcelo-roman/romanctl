from collections.abc import Callable, Sequence
from datetime import datetime
from typing import Protocol

from sample_app.domain.orders.order import Order
from sample_app.domain.shared.events import DomainEvent

Clock = Callable[[], datetime]


class OrderRepository(Protocol):
    def save(self, order: Order) -> None: ...

    def find(self, order_id: str) -> Order | None: ...

    def list_recent(self, limit: int) -> list[Order]: ...


class EventPublisher(Protocol):
    def publish(self, events: Sequence[DomainEvent]) -> None: ...
