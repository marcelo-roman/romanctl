from rest_framework import serializers

from sample_items.models import NAME_MAX_LENGTH, SampleItem

DEFAULT_LIMIT = 50
MAX_LIMIT = 200


class SampleItemSerializer(serializers.ModelSerializer):
    class Meta:
        model = SampleItem
        fields = ["id", "name", "created_at"]
        read_only_fields = fields


class SampleItemInputSerializer(serializers.Serializer):
    name = serializers.CharField(max_length=NAME_MAX_LENGTH)


class SampleItemQuerySerializer(serializers.Serializer):
    limit = serializers.IntegerField(required=False, min_value=1, max_value=MAX_LIMIT, default=DEFAULT_LIMIT)
