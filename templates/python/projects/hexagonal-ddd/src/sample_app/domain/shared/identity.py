import uuid

from sample_app.domain.shared.errors import BusinessRuleError


def new_id() -> str:
    return str(uuid.uuid4())


def parse_id(value: str) -> str:
    try:
        return str(uuid.UUID(value))
    except ValueError:
        raise BusinessRuleError("Identificador inválido.") from None
