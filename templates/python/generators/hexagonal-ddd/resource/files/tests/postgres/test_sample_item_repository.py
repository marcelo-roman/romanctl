from datetime import UTC, datetime

from sample_app.adapters.postgres.database import Database
from sample_app.adapters.postgres.sample_items import PostgresSampleItemRepository
from sample_app.domain.sample_items.sample_item import SampleItem
from sample_app.domain.shared.identity import new_id


def test_sample_item_round_trip(database: Database) -> None:
    repository = PostgresSampleItemRepository(database)
    sample_item = SampleItem.create("Primeiro", datetime.now(UTC).replace(microsecond=0))
    repository.save(sample_item)
    sample_item.rename("Segundo")
    repository.save(sample_item)

    loaded = repository.find(sample_item.id)

    assert loaded is not None
    assert (loaded.name, loaded.created_at) == ("Segundo", sample_item.created_at)
    assert sample_item.id in [recent.id for recent in repository.list_recent(200)]
    assert repository.delete(sample_item.id)
    assert not repository.delete(sample_item.id)
    assert repository.find(new_id()) is None
