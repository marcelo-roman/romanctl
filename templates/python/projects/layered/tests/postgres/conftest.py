import os

import pytest
from django.db import connection


@pytest.fixture
def postgres() -> None:
    if not os.environ.get("DATABASE_URL"):
        pytest.skip("DATABASE_URL não definida: romanctl up sample-app --no-run cria o banco e o .env")
    assert connection.vendor == "postgresql"
