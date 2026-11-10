from dataclasses import dataclass
from datetime import datetime

from sample_app.domain.sample_items.sample_item import SampleItem


@dataclass(frozen=True)
class SampleItemView:
    id: str
    name: str
    created_at: datetime

    @classmethod
    def of(cls, sample_item: SampleItem) -> "SampleItemView":
        return cls(sample_item.id, sample_item.name, sample_item.created_at)
