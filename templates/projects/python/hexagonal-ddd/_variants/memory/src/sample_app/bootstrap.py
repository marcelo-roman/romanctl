import logging
from datetime import UTC, datetime

from fastapi import FastAPI

from sample_app.adapters.http.app import create_app
from sample_app.adapters.http.orders import orders_router
from sample_app.adapters.logging.publisher import LoggingEventPublisher
from sample_app.adapters.memory.orders import InMemoryOrderRepository
from sample_app.application.orders.service import OrderService
# romanctl:imports


def utc_now() -> datetime:
    return datetime.now(UTC)


def build_app() -> FastAPI:
    events = LoggingEventPublisher(logging.getLogger("sample_app.events"))
    routers = [
        orders_router(OrderService(InMemoryOrderRepository(), events, utc_now)),
        # romanctl:routers
    ]
    return create_app(routers)
