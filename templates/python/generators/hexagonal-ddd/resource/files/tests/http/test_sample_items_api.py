from datetime import UTC, datetime

import pytest
from fastapi.testclient import TestClient

from sample_app.adapters.http.app import create_app
from sample_app.adapters.http.sample_items import sample_items_router
from sample_app.adapters.memory.sample_items import InMemorySampleItemRepository
from sample_app.application.sample_items.service import SampleItemService


@pytest.fixture
def client() -> TestClient:
    service = SampleItemService(InMemorySampleItemRepository(), lambda: datetime.now(UTC))
    return TestClient(create_app([sample_items_router(service)]))


def test_sample_item_crud(client: TestClient) -> None:
    response = client.post("/sample-items", json={"name": "Primeiro"})
    assert response.status_code == 201
    sample_item_id = response.json()["id"]
    assert response.headers["location"] == f"/sample-items/{sample_item_id}"

    assert client.put(f"/sample-items/{sample_item_id}", json={"name": "Segundo"}).status_code == 204
    assert client.get(f"/sample-items/{sample_item_id}").json()["name"] == "Segundo"
    assert [item["id"] for item in client.get("/sample-items").json()] == [sample_item_id]

    assert client.delete(f"/sample-items/{sample_item_id}").status_code == 204
    assert client.get(f"/sample-items/{sample_item_id}").status_code == 404


def test_sample_item_errors(client: TestClient) -> None:
    assert client.post("/sample-items", json={"name": " "}).status_code == 422
    assert client.post("/sample-items", json={}).status_code == 400
    assert client.delete("/sample-items/00000000-0000-4000-8000-000000000000").status_code == 404
