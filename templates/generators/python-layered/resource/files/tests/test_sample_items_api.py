import pytest
from rest_framework.test import APIClient

pytestmark = pytest.mark.django_db

PROBLEM_JSON = "application/problem+json"


def test_sample_item_crud(api_client: APIClient) -> None:
    response = api_client.post("/sample-items", {"name": " Primeiro "}, format="json")
    assert response.status_code == 201
    sample_item = response.json()
    assert response["Location"] == f"/sample-items/{sample_item['id']}"
    assert sample_item["name"] == "Primeiro"

    renamed = api_client.patch(f"/sample-items/{sample_item['id']}", {"name": "Segundo"}, format="json")
    assert renamed.status_code == 200
    assert api_client.get(f"/sample-items/{sample_item['id']}").json()["name"] == "Segundo"
    assert [item["id"] for item in api_client.get("/sample-items?limit=10").json()] == [sample_item["id"]]

    assert api_client.delete(f"/sample-items/{sample_item['id']}").status_code == 204
    assert api_client.get(f"/sample-items/{sample_item['id']}").status_code == 404


@pytest.mark.parametrize("body", [{}, {"name": "   "}, {"name": "x" * 201}])
def test_invalid_body_is_400(api_client: APIClient, body: dict) -> None:
    response = api_client.post("/sample-items", body, format="json")

    assert response.status_code == 400
    assert response["Content-Type"] == PROBLEM_JSON


def test_missing_sample_item_is_404(api_client: APIClient) -> None:
    response = api_client.delete("/sample-items/999999")

    assert response.status_code == 404
    assert response.json()["status"] == 404


def test_invalid_limit_is_400(api_client: APIClient) -> None:
    assert api_client.get("/sample-items?limit=0").status_code == 400
