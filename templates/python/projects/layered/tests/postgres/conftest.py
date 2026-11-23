import os

import pytest
from django.db import connection


@pytest.fixture
def postgres() -> None:
    if not os.environ.get("DATABASE_URL"):
        pytest.skip("DATABASE_URL not set: romanctl up --no-run creates the database and the .env")
    assert connection.vendor == "postgresql"
