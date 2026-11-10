from dataclasses import dataclass
from datetime import datetime
from decimal import Decimal

from sample_app.domain.orders.order import Order
from sample_app.domain.orders.order_item import OrderItem


@dataclass(frozen=True)
class OrderItemView:
    product_name: str
    quantity: int
    unit_price: Decimal
    subtotal: Decimal

    @classmethod
    def of(cls, item: OrderItem) -> "OrderItemView":
        return cls(item.product_name, item.quantity, item.unit_price.amount, item.subtotal.amount)


@dataclass(frozen=True)
class OrderView:
    id: str
    customer_name: str
    status: str
    created_at: datetime
    total: Decimal
    currency: str
    items: tuple[OrderItemView, ...]

    @classmethod
    def of(cls, order: Order) -> "OrderView":
        total = order.total
        return cls(
            id=order.id,
            customer_name=order.customer_name,
            status=str(order.status),
            created_at=order.created_at,
            total=total.amount,
            currency=total.currency,
            items=tuple(OrderItemView.of(item) for item in order.items),
        )
