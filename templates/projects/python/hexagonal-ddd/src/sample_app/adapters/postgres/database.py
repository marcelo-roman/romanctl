from collections.abc import AsyncIterator, Iterator
from contextlib import asynccontextmanager, contextmanager
from importlib.resources import files

from psycopg import Connection
from psycopg_pool import ConnectionPool

SCHEMA = files(__package__).joinpath("schema.sql").read_text(encoding="utf-8")


class Database:
    def __init__(self, url: str) -> None:
        self._pool = ConnectionPool(url, min_size=1, max_size=10, open=False)

    def open(self) -> None:
        self._pool.open(wait=True, timeout=10)
        with self._pool.connection() as connection:
            connection.execute(SCHEMA)

    def close(self) -> None:
        self._pool.close()

    @contextmanager
    def connection(self) -> Iterator[Connection]:
        with self._pool.connection() as connection:
            yield connection

    @asynccontextmanager
    async def lifespan(self, _: object) -> AsyncIterator[None]:
        self.open()
        try:
            yield
        finally:
            self.close()
