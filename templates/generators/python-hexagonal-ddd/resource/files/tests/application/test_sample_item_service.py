from datetime import UTC, datetime

import pytest

from sample_app.application.sample_items.errors import SampleItemNotFoundError
from sample_app.application.sample_items.service import SampleItemService
from sample_app.domain.sample_items.sample_item import SampleItem
from sample_app.domain.shared.identity import new_id


class FakeSampleItemRepository:
    def __init__(self) -> None:
        self.saved: dict[str, SampleItem] = {}

    def save(self, sample_item: SampleItem) -> None:
        self.saved[sample_item.id] = sample_item

    def find(self, sample_item_id: str) -> SampleItem | None:
        return self.saved.get(sample_item_id)

    def list_recent(self, limit: int) -> list[SampleItem]:
        return list(self.saved.values())[:limit]

    def delete(self, sample_item_id: str) -> bool:
        return self.saved.pop(sample_item_id, None) is not None


@pytest.fixture
def service() -> SampleItemService:
    return SampleItemService(FakeSampleItemRepository(), lambda: datetime(2026, 1, 1, tzinfo=UTC))


def test_create_rename_and_delete(service: SampleItemService) -> None:
    sample_item_id = service.create("Primeiro")
    service.rename(sample_item_id, "Segundo")

    assert service.get(sample_item_id).name == "Segundo"
    assert [view.id for view in service.list_recent()] == [sample_item_id]

    service.delete(sample_item_id)
    with pytest.raises(SampleItemNotFoundError):
        service.get(sample_item_id)


def test_unknown_sample_item_is_not_found(service: SampleItemService) -> None:
    with pytest.raises(SampleItemNotFoundError):
        service.rename(new_id(), "Nome")
    with pytest.raises(SampleItemNotFoundError):
        service.delete(new_id())
