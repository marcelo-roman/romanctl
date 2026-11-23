from sample_app.errors import NotFoundError
from tasks.models import Task


def list_tasks(*, limit: int, done: bool | None = None) -> list[Task]:
    tasks = Task.objects.all()
    if done is not None:
        tasks = tasks.filter(done=done)
    return list(tasks[:limit])


def get_task(task_id: int) -> Task:
    task = Task.objects.filter(pk=task_id).first()
    if task is None:
        raise NotFoundError(f"Task {task_id} not found.")
    return task
