from threading import Lock

from sample_app.domain.sample_items.sample_item import SampleItem


class InMemorySampleItemRepository:
    def __init__(self) -> None:
        self._sample_items: dict[str, SampleItem] = {}
        self._lock = Lock()

    def save(self, sample_item: SampleItem) -> None:
        with self._lock:
            self._sample_items[sample_item.id] = _copy(sample_item)

    def find(self, sample_item_id: str) -> SampleItem | None:
        with self._lock:
            sample_item = self._sample_items.get(sample_item_id)
        if sample_item is None:
            return None
        return _copy(sample_item)

    def list_recent(self, limit: int) -> list[SampleItem]:
        with self._lock:
            sample_items = sorted(self._sample_items.values(), key=lambda item: item.created_at, reverse=True)
        return [_copy(sample_item) for sample_item in sample_items[:limit]]

    def delete(self, sample_item_id: str) -> bool:
        with self._lock:
            return self._sample_items.pop(sample_item_id, None) is not None


def _copy(sample_item: SampleItem) -> SampleItem:
    return SampleItem(sample_item.id, sample_item.name, sample_item.created_at)
