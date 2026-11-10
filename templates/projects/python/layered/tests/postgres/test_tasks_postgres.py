import pytest
from django.core.management import call_command
from django.db import connection
from django.utils import timezone

from tasks.models import Task
from tasks.selectors import get_task, list_tasks
from tasks.services import create_task, update_task


@pytest.mark.django_db(transaction=True)
def test_migrations_apply_and_revert(postgres: None) -> None:
    call_command("migrate", "tasks", "zero", verbosity=0)
    assert Task._meta.db_table not in connection.introspection.table_names()

    call_command("migrate", "tasks", verbosity=0)
    assert Task._meta.db_table in connection.introspection.table_names()


@pytest.mark.django_db
def test_task_round_trip(postgres: None) -> None:
    task = create_task(title="  Comprar pão ", due_date=timezone.localdate())

    loaded = get_task(task.id)
    assert loaded.title == "Comprar pão"
    assert loaded.due_date == timezone.localdate()
    assert loaded.created_at == task.created_at

    update_task(loaded, {"done": True})

    assert [item.id for item in list_tasks(limit=200, done=True)] == [task.id]
