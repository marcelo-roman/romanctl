from collections.abc import Iterable
from datetime import datetime

from sample_app.domain.orders.events import OrderPlaced
from sample_app.domain.orders.money import DEFAULT_CURRENCY, Money
from sample_app.domain.orders.order_item import OrderItem
from sample_app.domain.orders.status import OrderStatus
from sample_app.domain.shared.errors import BusinessRuleError
from sample_app.domain.shared.events import DomainEvent
from sample_app.domain.shared.identity import new_id


class Order:
    def __init__(
        self,
        id: str,
        customer_name: str,
        status: OrderStatus,
        created_at: datetime,
        items: Iterable[OrderItem] = (),
    ) -> None:
        self._id = id
        self._customer_name = customer_name
        self._status = status
        self._created_at = created_at
        self._items = list(items)
        self._events: list[DomainEvent] = []

    @classmethod
    def create(cls, customer_name: str, now: datetime) -> "Order":
        if not customer_name.strip():
            raise BusinessRuleError("O nome do cliente é obrigatório.")
        return cls(new_id(), customer_name.strip(), OrderStatus.DRAFT, now)

    @property
    def id(self) -> str:
        return self._id

    @property
    def customer_name(self) -> str:
        return self._customer_name

    @property
    def status(self) -> OrderStatus:
        return self._status

    @property
    def created_at(self) -> datetime:
        return self._created_at

    @property
    def items(self) -> tuple[OrderItem, ...]:
        return tuple(self._items)

    @property
    def currency(self) -> str:
        if not self._items:
            return DEFAULT_CURRENCY
        return self._items[0].unit_price.currency

    @property
    def total(self) -> Money:
        total = Money.zero(self.currency)
        for item in self._items:
            total = total.add(item.subtotal)
        return total

    def add_item(self, product_name: str, quantity: int, unit_price: Money) -> None:
        if self._status != OrderStatus.DRAFT:
            raise BusinessRuleError("Um pedido já realizado não pode ser alterado.")
        if self._items and unit_price.currency != self.currency:
            raise BusinessRuleError("Todos os itens do pedido devem usar a mesma moeda.")
        self._items.append(OrderItem.create(product_name, quantity, unit_price))

    def place(self, now: datetime) -> None:
        if self._status != OrderStatus.DRAFT:
            raise BusinessRuleError("O pedido já foi realizado.")
        if not self._items:
            raise BusinessRuleError("Um pedido sem itens não pode ser realizado.")
        self._status = OrderStatus.PLACED
        total = self.total
        self._events.append(OrderPlaced(self._id, total.cents, total.currency, now))

    def pull_events(self) -> list[DomainEvent]:
        events = self._events
        self._events = []
        return events
