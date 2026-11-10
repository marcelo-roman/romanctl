from threading import Lock

from sample_app.domain.orders.order import Order


class InMemoryOrderRepository:
    def __init__(self) -> None:
        self._orders: dict[str, Order] = {}
        self._lock = Lock()

    def save(self, order: Order) -> None:
        with self._lock:
            self._orders[order.id] = _copy(order)

    def find(self, order_id: str) -> Order | None:
        with self._lock:
            order = self._orders.get(order_id)
        if order is None:
            return None
        return _copy(order)

    def list_recent(self, limit: int) -> list[Order]:
        with self._lock:
            orders = sorted(self._orders.values(), key=lambda order: order.created_at, reverse=True)
        return [_copy(order) for order in orders[:limit]]


def _copy(order: Order) -> Order:
    return Order(order.id, order.customer_name, order.status, order.created_at, order.items)
