from datetime import datetime

from sample_app.domain.shared.errors import BusinessRuleError
from sample_app.domain.shared.identity import new_id


class SampleItem:
    def __init__(self, id: str, name: str, created_at: datetime) -> None:
        self._id = id
        self._name = name
        self._created_at = created_at

    @classmethod
    def create(cls, name: str, now: datetime) -> "SampleItem":
        return cls(new_id(), _valid_name(name), now)

    @property
    def id(self) -> str:
        return self._id

    @property
    def name(self) -> str:
        return self._name

    @property
    def created_at(self) -> datetime:
        return self._created_at

    def rename(self, name: str) -> None:
        self._name = _valid_name(name)


def _valid_name(name: str) -> str:
    if not name.strip():
        raise BusinessRuleError("O nome é obrigatório.")
    return name.strip()
