from collections.abc import Mapping
from datetime import date
from typing import Any

from django.utils import timezone

from sample_app.errors import BusinessRuleError
from tasks.models import TITLE_MAX_LENGTH, Task


def create_task(*, title: str, due_date: date | None = None) -> Task:
    return Task.objects.create(title=_valid_title(title), due_date=_valid_due_date(due_date))


def update_task(task: Task, changes: Mapping[str, Any]) -> Task:
    if "title" in changes:
        task.title = _valid_title(changes["title"])
    if "due_date" in changes and changes["due_date"] != task.due_date:
        task.due_date = _valid_due_date(changes["due_date"])
    if "done" in changes:
        task.done = bool(changes["done"])
    task.save()
    return task


def delete_task(task: Task) -> None:
    task.delete()


def _valid_title(title: str) -> str:
    trimmed = title.strip()
    if not trimmed:
        raise BusinessRuleError("O título é obrigatório.")
    if len(trimmed) > TITLE_MAX_LENGTH:
        raise BusinessRuleError(f"O título passa de {TITLE_MAX_LENGTH} caracteres.")
    return trimmed


def _valid_due_date(due_date: date | None) -> date | None:
    if due_date is not None and due_date < timezone.localdate():
        raise BusinessRuleError("A data de vencimento não pode estar no passado.")
    return due_date
