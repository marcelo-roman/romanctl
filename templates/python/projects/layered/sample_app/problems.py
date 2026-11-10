import logging
from typing import Any

from django.http import HttpRequest, JsonResponse
from rest_framework.response import Response
from rest_framework.views import exception_handler

from sample_app.errors import BusinessRuleError, NotFoundError

PROBLEM_JSON = "application/problem+json"
TITLES = {
    400: "Requisição inválida",
    404: "Não encontrado",
    405: "Método não permitido",
    406: "Formato de resposta não aceito",
    415: "Tipo de conteúdo não suportado",
    422: "Regra de negócio violada",
    500: "Erro interno",
}
UNWRAPPED_KEYS = {"detail", "non_field_errors"}

logger = logging.getLogger(__name__)


def problem_body(status: int, detail: str = "") -> dict[str, Any]:
    return {"status": status, "title": TITLES.get(status, "Erro na requisição"), "detail": detail}


def problem_exception_handler(error: Exception, context: dict[str, Any]) -> Response:
    if isinstance(error, BusinessRuleError):
        return _problem(422, str(error))
    if isinstance(error, NotFoundError):
        return _problem(404, str(error))
    response = exception_handler(error, context)
    if response is None:
        logger.error("request failed", exc_info=error)
        return _problem(500)
    headers = {name: value for name, value in response.items() if name.lower() != "content-type"}
    return _problem(response.status_code, "; ".join(_messages(response.data)), headers)


def page_not_found(request: HttpRequest, exception: Exception) -> JsonResponse:
    return _plain_problem(404, f"Nada em {request.path}.")


def server_error(request: HttpRequest) -> JsonResponse:
    return _plain_problem(500)


def _plain_problem(status: int, detail: str = "") -> JsonResponse:
    body = problem_body(status, detail)
    return JsonResponse(body, status=status, content_type=PROBLEM_JSON, json_dumps_params={"ensure_ascii": False})


def _problem(status: int, detail: str = "", headers: dict[str, str] | None = None) -> Response:
    return Response(problem_body(status, detail), status=status, headers=headers, content_type=PROBLEM_JSON)


def _messages(data: Any, field: str = "") -> list[str]:
    if isinstance(data, dict):
        return [message for key, value in data.items() for message in _messages(value, _field(field, str(key)))]
    if isinstance(data, list):
        return [message for item in data for message in _messages(item, field)]
    if not field:
        return [str(data)]
    return [f"{field}: {data}"]


def _field(parent: str, key: str) -> str:
    if key in UNWRAPPED_KEYS:
        return parent
    if not parent:
        return key
    return f"{parent}.{key}"
