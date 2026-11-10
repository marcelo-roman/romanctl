from rest_framework import status
from rest_framework.request import Request
from rest_framework.response import Response
from rest_framework.views import APIView

from sample_items.selectors import get_sample_item, list_sample_items
from sample_items.serializers import SampleItemInputSerializer, SampleItemQuerySerializer, SampleItemSerializer
from sample_items.services import create_sample_item, delete_sample_item, rename_sample_item


class SampleItemListView(APIView):
    def get(self, request: Request) -> Response:
        query = SampleItemQuerySerializer(data=request.query_params.dict())
        query.is_valid(raise_exception=True)
        sample_items = list_sample_items(**query.validated_data)
        return Response(SampleItemSerializer(sample_items, many=True).data)

    def post(self, request: Request) -> Response:
        body = SampleItemInputSerializer(data=request.data)
        body.is_valid(raise_exception=True)
        sample_item = create_sample_item(**body.validated_data)
        headers = {"Location": f"/sample-items/{sample_item.id}"}
        return Response(SampleItemSerializer(sample_item).data, status=status.HTTP_201_CREATED, headers=headers)


class SampleItemDetailView(APIView):
    def get(self, request: Request, sample_item_id: int) -> Response:
        return Response(SampleItemSerializer(get_sample_item(sample_item_id)).data)

    def patch(self, request: Request, sample_item_id: int) -> Response:
        sample_item = get_sample_item(sample_item_id)
        body = SampleItemInputSerializer(data=request.data)
        body.is_valid(raise_exception=True)
        return Response(SampleItemSerializer(rename_sample_item(sample_item, **body.validated_data)).data)

    def delete(self, request: Request, sample_item_id: int) -> Response:
        delete_sample_item(get_sample_item(sample_item_id))
        return Response(status=status.HTTP_204_NO_CONTENT)
