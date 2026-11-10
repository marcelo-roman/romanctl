import logging
import os
from datetime import UTC, datetime

from fastapi import FastAPI

from sample_app.adapters.http.app import create_app
from sample_app.adapters.http.orders import orders_router
from sample_app.adapters.logging.publisher import LoggingEventPublisher
from sample_app.adapters.postgres.database import Database
from sample_app.adapters.postgres.orders import PostgresOrderRepository
from sample_app.application.orders.service import OrderService
# romanctl:imports

DEFAULT_DATABASE_URL = "postgres://postgres:postgres@127.0.0.1:5432/sample_app"


def utc_now() -> datetime:
    return datetime.now(UTC)


def build_app() -> FastAPI:
    database = Database(os.environ.get("DATABASE_URL", DEFAULT_DATABASE_URL))
    events = LoggingEventPublisher(logging.getLogger("sample_app.events"))
    routers = [
        orders_router(OrderService(PostgresOrderRepository(database), events, utc_now)),
        # romanctl:routers
    ]
    return create_app(routers, lifespan=database.lifespan)
