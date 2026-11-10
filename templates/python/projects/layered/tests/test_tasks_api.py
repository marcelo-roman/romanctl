from datetime import timedelta

import pytest
from django.utils import timezone
from rest_framework.test import APIClient

from tasks.models import Task

pytestmark = pytest.mark.django_db

PROBLEM_JSON = "application/problem+json"


def create(api_client: APIClient, title: str = "Comprar pão", **fields: object) -> dict:
    response = api_client.post("/tasks", {"title": title, **fields}, format="json")
    assert response.status_code == 201, response.content
    return response.json()


def assert_problem(response, status: int) -> dict:
    assert response.status_code == status, response.content
    assert response["Content-Type"] == PROBLEM_JSON
    body = response.json()
    assert body["status"] == status
    assert body["title"]
    return body


def test_health(api_client: APIClient) -> None:
    response = api_client.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_task_lifecycle(api_client: APIClient) -> None:
    due = (timezone.localdate() + timedelta(days=3)).isoformat()
    response = api_client.post("/tasks", {"title": "  Comprar pão  ", "due_date": due}, format="json")

    assert response.status_code == 201
    task = response.json()
    assert response["Location"] == f"/tasks/{task['id']}"
    assert set(task) == {"id", "title", "done", "due_date", "created_at"}
    assert task["title"] == "Comprar pão"
    assert task["done"] is False
    assert task["due_date"] == due

    assert api_client.get(f"/tasks/{task['id']}").json() == task

    patched = api_client.patch(f"/tasks/{task['id']}", {"done": True, "due_date": None}, format="json")
    assert patched.status_code == 200
    assert patched.json()["done"] is True
    assert patched.json()["due_date"] is None
    assert patched.json()["title"] == "Comprar pão"

    assert api_client.delete(f"/tasks/{task['id']}").status_code == 204
    assert_problem(api_client.get(f"/tasks/{task['id']}"), 404)


def test_list_filters_and_limits(api_client: APIClient) -> None:
    first = create(api_client, "Primeira")
    second = create(api_client, "Segunda")
    api_client.patch(f"/tasks/{first['id']}", {"done": True}, format="json")

    assert [task["id"] for task in api_client.get("/tasks").json()] == [second["id"], first["id"]]
    assert [task["id"] for task in api_client.get("/tasks?done=true").json()] == [first["id"]]
    assert [task["id"] for task in api_client.get("/tasks?done=false").json()] == [second["id"]]
    assert [task["id"] for task in api_client.get("/tasks?done=").json()] == [second["id"], first["id"]]
    assert [task["id"] for task in api_client.get("/tasks?limit=1").json()] == [second["id"]]


def test_list_returns_fifty_by_default(api_client: APIClient) -> None:
    Task.objects.bulk_create(Task(title=f"Tarefa {number}") for number in range(55))

    assert len(api_client.get("/tasks").json()) == 50
    assert len(api_client.get("/tasks?limit=200").json()) == 55


@pytest.mark.parametrize("query", ["limit=0", "limit=201", "limit=muitas", "done=talvez"])
def test_invalid_query_is_400(api_client: APIClient, query: str) -> None:
    assert_problem(api_client.get(f"/tasks?{query}"), 400)


@pytest.mark.parametrize("body", [{}, {"title": "   "}, {"title": "x" * 201}, {"title": "Ok", "due_date": "ontem"}])
def test_invalid_body_is_400(api_client: APIClient, body: dict) -> None:
    problem = assert_problem(api_client.post("/tasks", body, format="json"), 400)

    assert problem["detail"]


def test_missing_title_names_the_field(api_client: APIClient) -> None:
    problem = assert_problem(api_client.post("/tasks", {}, format="json"), 400)

    assert problem["detail"].startswith("title: ")


def test_malformed_json_is_400(api_client: APIClient) -> None:
    assert_problem(api_client.post("/tasks", "{", content_type="application/json"), 400)


def test_due_date_in_the_past_is_422(api_client: APIClient) -> None:
    yesterday = (timezone.localdate() - timedelta(days=1)).isoformat()

    response = api_client.post("/tasks", {"title": "Atrasada", "due_date": yesterday}, format="json")

    problem = assert_problem(response, 422)

    assert problem["title"] == "Regra de negócio violada"
    assert "passado" in problem["detail"]


def test_patch_validates_the_body(api_client: APIClient) -> None:
    task = create(api_client)

    assert_problem(api_client.patch(f"/tasks/{task['id']}", {"done": "talvez"}, format="json"), 400)
    assert_problem(api_client.patch(f"/tasks/{task['id']}", {"title": ""}, format="json"), 400)


def test_missing_task_is_404(api_client: APIClient) -> None:
    for response in [
        api_client.get("/tasks/999999"),
        api_client.patch("/tasks/999999", {"done": True}, format="json"),
        api_client.delete("/tasks/999999"),
    ]:
        assert "999999" in assert_problem(response, 404)["detail"]


def test_unknown_route_is_404(api_client: APIClient) -> None:
    assert_problem(api_client.get("/tasks/abc"), 404)


def test_unsupported_method_is_405(api_client: APIClient) -> None:
    task = create(api_client)

    response = api_client.put(f"/tasks/{task['id']}", {"title": "Outra"}, format="json")

    assert_problem(response, 405)
    assert "PATCH" in response["Allow"]
