from datetime import datetime

from fastapi import APIRouter, Response

from sample_app.adapters.http.schemas import CamelModel, CreatedResponse
from sample_app.application.sample_items.service import SampleItemService
from sample_app.application.sample_items.views import SampleItemView


class SampleItemRequest(CamelModel):
    name: str


class SampleItemResponse(CamelModel):
    id: str
    name: str
    created_at: datetime

    @classmethod
    def of(cls, view: SampleItemView) -> "SampleItemResponse":
        return cls(id=view.id, name=view.name, created_at=view.created_at)


def sample_items_router(service: SampleItemService) -> APIRouter:
    router = APIRouter(prefix="/sample-items", tags=["sample-items"])

    @router.post("", status_code=201)
    def create_sample_item(body: SampleItemRequest, response: Response) -> CreatedResponse:
        sample_item_id = service.create(body.name)
        response.headers["Location"] = f"/sample-items/{sample_item_id}"
        return CreatedResponse(id=sample_item_id)

    @router.get("")
    def list_sample_items(limit: int = 0) -> list[SampleItemResponse]:
        return [SampleItemResponse.of(view) for view in service.list_recent(limit)]

    @router.get("/{sample_item_id}")
    def get_sample_item(sample_item_id: str) -> SampleItemResponse:
        return SampleItemResponse.of(service.get(sample_item_id))

    @router.put("/{sample_item_id}", status_code=204)
    def rename_sample_item(sample_item_id: str, body: SampleItemRequest) -> None:
        service.rename(sample_item_id, body.name)

    @router.delete("/{sample_item_id}", status_code=204)
    def delete_sample_item(sample_item_id: str) -> None:
        service.delete(sample_item_id)

    return router
