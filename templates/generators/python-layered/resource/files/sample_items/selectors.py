from sample_app.errors import NotFoundError
from sample_items.models import SampleItem


def list_sample_items(*, limit: int) -> list[SampleItem]:
    return list(SampleItem.objects.all()[:limit])


def get_sample_item(sample_item_id: int) -> SampleItem:
    sample_item = SampleItem.objects.filter(pk=sample_item_id).first()
    if sample_item is None:
        raise NotFoundError(f"Sample Item {sample_item_id} não encontrado.")
    return sample_item
