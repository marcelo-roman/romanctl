from dataclasses import dataclass
from datetime import datetime
from typing import ClassVar

from sample_app.domain.shared.events import DomainEvent


@dataclass(frozen=True)
class OrderPlaced(DomainEvent):
    name: ClassVar[str] = "order.placed"

    order_id: str
    total_cents: int
    currency: str
    occurred_at: datetime
