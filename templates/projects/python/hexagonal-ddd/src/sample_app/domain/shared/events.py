from dataclasses import dataclass
from typing import ClassVar


@dataclass(frozen=True)
class DomainEvent:
    name: ClassVar[str] = "domain.event"
