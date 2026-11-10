import pytest

from sample_items.selectors import get_sample_item, list_sample_items
from sample_items.services import create_sample_item, rename_sample_item


@pytest.mark.django_db
def test_sample_item_round_trip(postgres: None) -> None:
    sample_item = create_sample_item(name="Primeiro")
    rename_sample_item(sample_item, name="Segundo")

    loaded = get_sample_item(sample_item.id)

    assert loaded.name == "Segundo"
    assert loaded.created_at == sample_item.created_at
    assert sample_item.id in [item.id for item in list_sample_items(limit=200)]
