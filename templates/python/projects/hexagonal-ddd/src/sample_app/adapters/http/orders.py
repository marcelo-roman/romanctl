from datetime import datetime
from decimal import Decimal

from fastapi import APIRouter, Response

from sample_app.adapters.http.schemas import CamelModel, CreatedResponse
from sample_app.application.orders.commands import AddItem
from sample_app.application.orders.service import OrderService
from sample_app.application.orders.views import OrderItemView, OrderView


class CreateOrderRequest(CamelModel):
    customer_name: str


class AddItemRequest(CamelModel):
    product_name: str
    quantity: int
    unit_price: Decimal
    currency: str = ""

    def to_command(self) -> AddItem:
        return AddItem(self.product_name, self.quantity, self.unit_price, self.currency)


class OrderItemResponse(CamelModel):
    product_name: str
    quantity: int
    unit_price: float
    subtotal: float

    @classmethod
    def of(cls, view: OrderItemView) -> "OrderItemResponse":
        return cls(
            product_name=view.product_name,
            quantity=view.quantity,
            unit_price=float(view.unit_price),
            subtotal=float(view.subtotal),
        )


class OrderResponse(CamelModel):
    id: str
    customer_name: str
    status: str
    created_at: datetime
    total: float
    currency: str
    items: list[OrderItemResponse]

    @classmethod
    def of(cls, view: OrderView) -> "OrderResponse":
        return cls(
            id=view.id,
            customer_name=view.customer_name,
            status=view.status,
            created_at=view.created_at,
            total=float(view.total),
            currency=view.currency,
            items=[OrderItemResponse.of(item) for item in view.items],
        )


def orders_router(service: OrderService) -> APIRouter:
    router = APIRouter(prefix="/orders", tags=["orders"])

    @router.post("", status_code=201)
    def create_order(body: CreateOrderRequest, response: Response) -> CreatedResponse:
        order_id = service.create(body.customer_name)
        response.headers["Location"] = f"/orders/{order_id}"
        return CreatedResponse(id=order_id)

    @router.post("/{order_id}/items", status_code=204)
    def add_item(order_id: str, body: AddItemRequest) -> None:
        service.add_item(order_id, body.to_command())

    @router.post("/{order_id}/place", status_code=204)
    def place_order(order_id: str) -> None:
        service.place(order_id)

    @router.get("/{order_id}")
    def get_order(order_id: str) -> OrderResponse:
        return OrderResponse.of(service.get(order_id))

    @router.get("")
    def list_orders(limit: int = 0) -> list[OrderResponse]:
        return [OrderResponse.of(view) for view in service.list_recent(limit)]

    return router
