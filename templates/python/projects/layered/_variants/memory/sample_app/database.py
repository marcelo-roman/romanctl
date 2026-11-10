import os
from pathlib import Path
from typing import Any


def default_database(base_dir: Path) -> dict[str, Any]:
    return {"ENGINE": "django.db.backends.sqlite3", "NAME": os.environ.get("SQLITE_PATH", base_dir / "db.sqlite3")}
