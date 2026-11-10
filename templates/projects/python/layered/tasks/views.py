from rest_framework import status
from rest_framework.request import Request
from rest_framework.response import Response
from rest_framework.views import APIView

from tasks.selectors import get_task, list_tasks
from tasks.serializers import TaskCreateSerializer, TaskQuerySerializer, TaskSerializer, TaskUpdateSerializer
from tasks.services import create_task, delete_task, update_task


class TaskListView(APIView):
    def get(self, request: Request) -> Response:
        query = TaskQuerySerializer(data=request.query_params.dict())
        query.is_valid(raise_exception=True)
        tasks = list_tasks(**query.validated_data)
        return Response(TaskSerializer(tasks, many=True).data)

    def post(self, request: Request) -> Response:
        body = TaskCreateSerializer(data=request.data)
        body.is_valid(raise_exception=True)
        task = create_task(**body.validated_data)
        headers = {"Location": f"/tasks/{task.id}"}
        return Response(TaskSerializer(task).data, status=status.HTTP_201_CREATED, headers=headers)


class TaskDetailView(APIView):
    def get(self, request: Request, task_id: int) -> Response:
        return Response(TaskSerializer(get_task(task_id)).data)

    def patch(self, request: Request, task_id: int) -> Response:
        task = get_task(task_id)
        body = TaskUpdateSerializer(data=request.data)
        body.is_valid(raise_exception=True)
        return Response(TaskSerializer(update_task(task, body.validated_data)).data)

    def delete(self, request: Request, task_id: int) -> Response:
        delete_task(get_task(task_id))
        return Response(status=status.HTTP_204_NO_CONTENT)
