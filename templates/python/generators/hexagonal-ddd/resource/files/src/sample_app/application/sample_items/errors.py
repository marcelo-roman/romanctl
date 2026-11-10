from sample_app.domain.shared.errors import NotFoundError


class SampleItemNotFoundError(NotFoundError):
    def __init__(self) -> None:
        super().__init__("SampleItem não encontrado.")
