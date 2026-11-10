import logging
from collections.abc import Callable
from datetime import UTC, datetime, timedelta
from itertools import count

import pytest
from fastapi.testclient import TestClient

from sample_app.adapters.http.app import create_app
from sample_app.adapters.http.orders import orders_router
from sample_app.adapters.logging.publisher import LoggingEventPublisher
from sample_app.adapters.memory.orders import InMemoryOrderRepository
from sample_app.application.orders.service import OrderService

MISSING_ID = "00000000-0000-4000-8000-000000000000"


def ticking_clock() -> Callable[[], datetime]:
    ticks = count()
    start = datetime(2026, 1, 1, tzinfo=UTC)
    return lambda: start + timedelta(seconds=next(ticks))


@pytest.fixture
def client() -> TestClient:
    service = OrderService(InMemoryOrderRepository(), LoggingEventPublisher(logging.getLogger("test")), ticking_clock())
    return TestClient(create_app([orders_router(service)]))


def create_order(client: TestClient, customer: str = "Ana") -> str:
    response = client.post("/orders", json={"customerName": customer})
    assert response.status_code == 201
    return response.json()["id"]


def test_health(client: TestClient) -> None:
    response = client.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_order_lifecycle(client: TestClient) -> None:
    response = client.post("/orders", json={"customerName": "Ana"})
    assert response.status_code == 201
    order_id = response.json()["id"]
    assert response.headers["location"] == f"/orders/{order_id}"

    item = {"productName": "Caneta", "quantity": 2, "unitPrice": 3.5}
    assert client.post(f"/orders/{order_id}/items", json=item).status_code == 204
    assert client.post(f"/orders/{order_id}/place").status_code == 204

    order = client.get(f"/orders/{order_id}").json()
    assert order["status"] == "Placed"
    assert order["customerName"] == "Ana"
    assert order["total"] == 7
    assert order["currency"] == "BRL"
    assert order["items"] == [{"productName": "Caneta", "quantity": 2, "unitPrice": 3.5, "subtotal": 7}]


def test_list_returns_most_recent_first(client: TestClient) -> None:
    first = create_order(client, "Ana")
    second = create_order(client, "Bia")

    assert [order["id"] for order in client.get("/orders").json()] == [second, first]
    assert [order["id"] for order in client.get("/orders?limit=1").json()] == [second]


def test_business_rule_violation_is_422(client: TestClient) -> None:
    response = client.post("/orders", json={"customerName": ""})

    assert response.status_code == 422
    assert response.headers["content-type"] == "application/problem+json"
    assert response.json() == {"status": 422, "title": "Regra de negócio violada",
                               "detail": "O nome do cliente é obrigatório."}


def test_placing_an_empty_order_is_422(client: TestClient) -> None:
    order_id = create_order(client)

    assert client.post(f"/orders/{order_id}/place").status_code == 422


def test_missing_order_is_404(client: TestClient) -> None:
    response = client.get(f"/orders/{MISSING_ID}")

    assert response.status_code == 404
    assert response.json()["detail"] == "Pedido não encontrado."


@pytest.mark.parametrize("body", ["{", '{"quantity": 1}', '{"customerName": null}'])
def test_invalid_body_is_400(client: TestClient, body: str) -> None:
    response = client.post("/orders", content=body, headers={"content-type": "application/json"})

    assert response.status_code == 400
    assert response.json()["title"] == "Requisição inválida"
