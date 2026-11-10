from dataclasses import dataclass
from decimal import Decimal


@dataclass(frozen=True)
class AddItem:
    product_name: str
    quantity: int
    unit_price: Decimal
    currency: str = ""
