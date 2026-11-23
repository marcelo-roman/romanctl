from dataclasses import dataclass
from decimal import ROUND_HALF_UP, Decimal

from sample_app.domain.shared.errors import BusinessRuleError

DEFAULT_CURRENCY = "USD"
CENTS = Decimal(100)


@dataclass(frozen=True)
class Money:
    cents: int
    currency: str = DEFAULT_CURRENCY

    def __post_init__(self) -> None:
        if self.cents < 0:
            raise BusinessRuleError("The amount cannot be negative.")
        currency = self.currency.strip()
        if len(currency) != 3 or not currency.isalpha():
            raise BusinessRuleError("The currency must have 3 letters, like USD.")
        object.__setattr__(self, "currency", currency.upper())

    @classmethod
    def of(cls, amount: Decimal, currency: str = DEFAULT_CURRENCY) -> "Money":
        cents = (Decimal(amount) * CENTS).quantize(Decimal(1), rounding=ROUND_HALF_UP)
        return cls(int(cents), currency or DEFAULT_CURRENCY)

    @classmethod
    def zero(cls, currency: str = DEFAULT_CURRENCY) -> "Money":
        return cls(0, currency)

    @property
    def amount(self) -> Decimal:
        return Decimal(self.cents) / CENTS

    def multiply(self, quantity: int) -> "Money":
        return Money(self.cents * quantity, self.currency)

    def add(self, other: "Money") -> "Money":
        if other.currency != self.currency:
            raise BusinessRuleError("Cannot add different currencies.")
        return Money(self.cents + other.cents, self.currency)
