from sample_app.domain.shared.errors import NotFoundError


class OrderNotFoundError(NotFoundError):
    def __init__(self) -> None:
        super().__init__("Pedido não encontrado.")
