from collections.abc import Iterable

from fastapi import APIRouter, FastAPI
from starlette.types import Lifespan

from sample_app.adapters.http.health import health_router
from sample_app.adapters.http.problems import register_problems


def create_app(routers: Iterable[APIRouter], lifespan: Lifespan[FastAPI] | None = None) -> FastAPI:
    app = FastAPI(title="Sample App", lifespan=lifespan)
    register_problems(app)
    app.include_router(health_router())
    for router in routers:
        app.include_router(router)
    return app
