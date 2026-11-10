import pytest

from sample_app.errors import BusinessRuleError, NotFoundError
from sample_items.selectors import get_sample_item, list_sample_items
from sample_items.services import create_sample_item, delete_sample_item, rename_sample_item

pytestmark = pytest.mark.django_db


def test_create_and_rename_trim_the_name() -> None:
    sample_item = create_sample_item(name="  Primeiro  ")
    assert sample_item.name == "Primeiro"

    rename_sample_item(sample_item, name=" Segundo ")

    assert get_sample_item(sample_item.id).name == "Segundo"


@pytest.mark.parametrize("name", ["", "   ", "x" * 201])
def test_invalid_names_are_rejected(name: str) -> None:
    with pytest.raises(BusinessRuleError):
        create_sample_item(name=name)


def test_list_returns_newest_first() -> None:
    first = create_sample_item(name="Primeiro")
    second = create_sample_item(name="Segundo")

    assert [sample_item.id for sample_item in list_sample_items(limit=50)] == [second.id, first.id]
    assert [sample_item.id for sample_item in list_sample_items(limit=1)] == [second.id]


def test_delete_removes_it() -> None:
    sample_item = create_sample_item(name="Descartável")

    delete_sample_item(sample_item)

    with pytest.raises(NotFoundError):
        get_sample_item(sample_item.id)
