from datetime import date, timedelta

import pytest
from django.utils import timezone

from sample_app.errors import BusinessRuleError, NotFoundError
from tasks.models import Task
from tasks.selectors import get_task, list_tasks
from tasks.services import create_task, delete_task, update_task

pytestmark = pytest.mark.django_db


def yesterday() -> date:
    return timezone.localdate() - timedelta(days=1)


def test_create_trims_the_title_and_starts_pending() -> None:
    task = create_task(title="  Comprar pão  ", due_date=timezone.localdate())

    assert task.title == "Comprar pão"
    assert task.done is False
    assert task.due_date == timezone.localdate()
    assert get_task(task.id).title == "Comprar pão"


@pytest.mark.parametrize("title", ["", "   ", "x" * 201])
def test_create_rejects_invalid_titles(title: str) -> None:
    with pytest.raises(BusinessRuleError):
        create_task(title=title)


def test_create_rejects_a_due_date_in_the_past() -> None:
    with pytest.raises(BusinessRuleError, match="passado"):
        create_task(title="Pagar a conta", due_date=yesterday())

    assert Task.objects.count() == 0


def test_update_changes_only_what_was_sent() -> None:
    task = create_task(title="Estudar")

    update_task(task, {"done": True})
    update_task(task, {"title": " Estudar Django "})

    loaded = get_task(task.id)
    assert loaded.title == "Estudar Django"
    assert loaded.done is True
    assert loaded.due_date is None


def test_update_rejects_moving_the_due_date_to_the_past() -> None:
    task = create_task(title="Estudar")

    with pytest.raises(BusinessRuleError):
        update_task(task, {"due_date": yesterday()})


def test_update_keeps_an_overdue_date_that_did_not_change() -> None:
    task = create_task(title="Estudar")
    Task.objects.filter(pk=task.id).update(due_date=yesterday())
    overdue = get_task(task.id)

    update_task(overdue, {"done": True, "due_date": yesterday()})

    assert get_task(task.id).done is True


def test_list_filters_by_status_and_returns_newest_first() -> None:
    first = create_task(title="Primeira")
    second = create_task(title="Segunda")
    update_task(first, {"done": True})

    assert [task.id for task in list_tasks(limit=50)] == [second.id, first.id]
    assert [task.id for task in list_tasks(limit=50, done=True)] == [first.id]
    assert [task.id for task in list_tasks(limit=50, done=False)] == [second.id]
    assert [task.id for task in list_tasks(limit=1)] == [second.id]


def test_delete_removes_the_task() -> None:
    task = create_task(title="Descartável")

    delete_task(task)

    with pytest.raises(NotFoundError):
        get_task(task.id)
