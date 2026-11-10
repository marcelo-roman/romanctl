from collections.abc import Callable
from datetime import datetime

from sample_app.application.sample_items.errors import SampleItemNotFoundError
from sample_app.application.sample_items.ports import SampleItemRepository
from sample_app.application.sample_items.views import SampleItemView
from sample_app.domain.sample_items.sample_item import SampleItem
from sample_app.domain.shared.identity import parse_id

MAX_LIST_LIMIT = 200


class SampleItemService:
    def __init__(self, repository: SampleItemRepository, clock: Callable[[], datetime]) -> None:
        self._repository = repository
        self._clock = clock

    def create(self, name: str) -> str:
        sample_item = SampleItem.create(name, self._clock())
        self._repository.save(sample_item)
        return sample_item.id

    def rename(self, sample_item_id: str, name: str) -> None:
        sample_item = self._load(sample_item_id)
        sample_item.rename(name)
        self._repository.save(sample_item)

    def get(self, sample_item_id: str) -> SampleItemView:
        return SampleItemView.of(self._load(sample_item_id))

    def list_recent(self, limit: int = MAX_LIST_LIMIT) -> list[SampleItemView]:
        if limit <= 0 or limit > MAX_LIST_LIMIT:
            limit = MAX_LIST_LIMIT
        return [SampleItemView.of(sample_item) for sample_item in self._repository.list_recent(limit)]

    def delete(self, sample_item_id: str) -> None:
        if not self._repository.delete(parse_id(sample_item_id)):
            raise SampleItemNotFoundError()

    def _load(self, sample_item_id: str) -> SampleItem:
        sample_item = self._repository.find(parse_id(sample_item_id))
        if sample_item is None:
            raise SampleItemNotFoundError()
        return sample_item
