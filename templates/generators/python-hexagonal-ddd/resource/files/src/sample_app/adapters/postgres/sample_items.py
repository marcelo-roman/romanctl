from typing import Any

from sample_app.adapters.postgres.database import Database
from sample_app.domain.sample_items.sample_item import SampleItem

UPSERT = """
    INSERT INTO sample_items (id, name, created_at) VALUES (%s, %s, %s)
    ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
"""
SELECT = "SELECT id, name, created_at FROM sample_items"


class PostgresSampleItemRepository:
    def __init__(self, database: Database) -> None:
        self._database = database

    def save(self, sample_item: SampleItem) -> None:
        with self._database.connection() as connection:
            connection.execute(UPSERT, (sample_item.id, sample_item.name, sample_item.created_at))

    def find(self, sample_item_id: str) -> SampleItem | None:
        sample_items = self._query(f"{SELECT} WHERE id = %s", (sample_item_id,))
        return sample_items[0] if sample_items else None

    def list_recent(self, limit: int) -> list[SampleItem]:
        return self._query(f"{SELECT} ORDER BY created_at DESC LIMIT %s", (limit,))

    def delete(self, sample_item_id: str) -> bool:
        with self._database.connection() as connection:
            return connection.execute("DELETE FROM sample_items WHERE id = %s", (sample_item_id,)).rowcount > 0

    def _query(self, query: str, params: tuple[Any, ...]) -> list[SampleItem]:
        with self._database.connection() as connection:
            rows = connection.execute(query, params).fetchall()
        return [SampleItem(str(row_id), name, created_at) for row_id, name, created_at in rows]
