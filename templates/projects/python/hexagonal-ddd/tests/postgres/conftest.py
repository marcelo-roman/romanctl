import os
from collections.abc import Iterator

import pytest

from sample_app.adapters.postgres.database import Database


@pytest.fixture
def database() -> Iterator[Database]:
    url = os.environ.get("DATABASE_URL")
    if not url:
        pytest.skip("DATABASE_URL not set: romanctl up sample-app --no-run creates the database and the .env")
    database = Database(url)
    database.open()
    yield database
    database.close()
