from dataclasses import dataclass

from sample_app.domain.orders.money import Money
from sample_app.domain.shared.errors import BusinessRuleError
from sample_app.domain.shared.identity import new_id


@dataclass(frozen=True)
class OrderItem:
    id: str
    product_name: str
    quantity: int
    unit_price: Money

    @classmethod
    def create(cls, product_name: str, quantity: int, unit_price: Money) -> "OrderItem":
        if not product_name.strip():
            raise BusinessRuleError("O nome do produto é obrigatório.")
        if quantity <= 0:
            raise BusinessRuleError("A quantidade deve ser maior que zero.")
        return cls(new_id(), product_name.strip(), quantity, unit_price)

    @property
    def subtotal(self) -> Money:
        return self.unit_price.multiply(self.quantity)
