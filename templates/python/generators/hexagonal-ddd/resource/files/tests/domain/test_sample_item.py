from datetime import UTC, datetime

import pytest

from sample_app.domain.sample_items.sample_item import SampleItem
from sample_app.domain.shared.errors import BusinessRuleError

NOW = datetime(2026, 1, 1, tzinfo=UTC)


def test_create_requires_name() -> None:
    with pytest.raises(BusinessRuleError):
        SampleItem.create(" ", NOW)

    assert SampleItem.create("  Primeiro ", NOW).name == "Primeiro"


def test_rename_requires_name() -> None:
    sample_item = SampleItem.create("Primeiro", NOW)
    with pytest.raises(BusinessRuleError):
        sample_item.rename("")
    sample_item.rename("Segundo")

    assert sample_item.name == "Segundo"
