from collections.abc import Callable

from sample_app.application.orders.commands import AddItem
from sample_app.application.orders.errors import OrderNotFoundError
from sample_app.application.orders.ports import Clock, EventPublisher, OrderRepository
from sample_app.application.orders.views import OrderView
from sample_app.domain.orders.money import Money
from sample_app.domain.orders.order import Order
from sample_app.domain.shared.identity import parse_id

MAX_LIST_LIMIT = 200


class OrderService:
    def __init__(self, repository: OrderRepository, events: EventPublisher, clock: Clock) -> None:
        self._repository = repository
        self._events = events
        self._clock = clock

    def create(self, customer_name: str) -> str:
        order = Order.create(customer_name, self._clock())
        self._repository.save(order)
        return order.id

    def add_item(self, order_id: str, command: AddItem) -> None:
        price = Money.of(command.unit_price, command.currency)
        self._change(order_id, lambda order: order.add_item(command.product_name, command.quantity, price))

    def place(self, order_id: str) -> None:
        self._change(order_id, lambda order: order.place(self._clock()))

    def get(self, order_id: str) -> OrderView:
        return OrderView.of(self._load(order_id))

    def list_recent(self, limit: int = MAX_LIST_LIMIT) -> list[OrderView]:
        if limit <= 0 or limit > MAX_LIST_LIMIT:
            limit = MAX_LIST_LIMIT
        return [OrderView.of(order) for order in self._repository.list_recent(limit)]

    def _load(self, order_id: str) -> Order:
        order = self._repository.find(parse_id(order_id))
        if order is None:
            raise OrderNotFoundError()
        return order

    def _change(self, order_id: str, apply: Callable[[Order], None]) -> None:
        order = self._load(order_id)
        apply(order)
        self._repository.save(order)
        self._events.publish(order.pull_events())
