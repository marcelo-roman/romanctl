from sample_app.errors import BusinessRuleError
from sample_items.models import NAME_MAX_LENGTH, SampleItem


def create_sample_item(*, name: str) -> SampleItem:
    return SampleItem.objects.create(name=_valid_name(name))


def rename_sample_item(sample_item: SampleItem, *, name: str) -> SampleItem:
    sample_item.name = _valid_name(name)
    sample_item.save(update_fields=["name"])
    return sample_item


def delete_sample_item(sample_item: SampleItem) -> None:
    sample_item.delete()


def _valid_name(name: str) -> str:
    trimmed = name.strip()
    if not trimmed:
        raise BusinessRuleError("O nome é obrigatório.")
    if len(trimmed) > NAME_MAX_LENGTH:
        raise BusinessRuleError(f"O nome passa de {NAME_MAX_LENGTH} caracteres.")
    return trimmed
