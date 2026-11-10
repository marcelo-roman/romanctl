import logging

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from sample_app.adapters.http.schemas import Problem
from sample_app.domain.shared.errors import BusinessRuleError, NotFoundError

PROBLEM_JSON = "application/problem+json"

logger = logging.getLogger(__name__)


def problem(status: int, title: str, detail: str = "") -> JSONResponse:
    body = Problem(status=status, title=title, detail=detail)
    return JSONResponse(body.model_dump(), status_code=status, media_type=PROBLEM_JSON)


async def invalid_request(_: Request, error: Exception) -> JSONResponse:
    errors = error.errors() if isinstance(error, RequestValidationError) else []
    return problem(400, "Requisição inválida", "; ".join(_describe(item) for item in errors))


async def business_rule(_: Request, error: Exception) -> JSONResponse:
    return problem(422, "Regra de negócio violada", str(error))


async def not_found(_: Request, error: Exception) -> JSONResponse:
    return problem(404, "Não encontrado", str(error))


async def unexpected(_: Request, error: Exception) -> JSONResponse:
    logger.error("request failed", exc_info=error)
    return problem(500, "Erro interno")


def register_problems(app: FastAPI) -> None:
    app.add_exception_handler(RequestValidationError, invalid_request)
    app.add_exception_handler(BusinessRuleError, business_rule)
    app.add_exception_handler(NotFoundError, not_found)
    app.add_exception_handler(Exception, unexpected)


def _describe(item: dict) -> str:
    location = ".".join(str(part) for part in item.get("loc", ()) if part != "body")
    message = item.get("msg", "")
    if not location:
        return message
    return f"{location}: {message}"
