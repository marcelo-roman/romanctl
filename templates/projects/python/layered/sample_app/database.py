import os
from pathlib import Path
from typing import Any

import dj_database_url


def sqlite(base_dir: Path) -> dict[str, Any]:
    return {"ENGINE": "django.db.backends.sqlite3", "NAME": os.environ.get("SQLITE_PATH", base_dir / "db.sqlite3")}


def default_database(base_dir: Path) -> dict[str, Any]:
    url = os.environ.get("DATABASE_URL", "")
    if not url:
        return sqlite(base_dir)
    return dict(dj_database_url.parse(url, conn_max_age=60, conn_health_checks=True))
