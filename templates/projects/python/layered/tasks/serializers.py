from rest_framework import serializers

from tasks.models import TITLE_MAX_LENGTH, Task

DEFAULT_LIMIT = 50
MAX_LIMIT = 200


class TaskSerializer(serializers.ModelSerializer):
    class Meta:
        model = Task
        fields = ["id", "title", "done", "due_date", "created_at"]
        read_only_fields = fields


class TaskCreateSerializer(serializers.Serializer):
    title = serializers.CharField(max_length=TITLE_MAX_LENGTH)
    due_date = serializers.DateField(required=False, allow_null=True)


class TaskUpdateSerializer(serializers.Serializer):
    title = serializers.CharField(max_length=TITLE_MAX_LENGTH, required=False)
    done = serializers.BooleanField(required=False)
    due_date = serializers.DateField(required=False, allow_null=True)


class TaskQuerySerializer(serializers.Serializer):
    done = serializers.BooleanField(required=False, allow_null=True, default=None)
    limit = serializers.IntegerField(required=False, min_value=1, max_value=MAX_LIMIT, default=DEFAULT_LIMIT)
