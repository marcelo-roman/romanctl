using SampleApp.Domain.Common;

namespace SampleApp.Domain.Orders;

public sealed record Money
{
    public decimal Amount { get; }
    public string Currency { get; }

    public Money(decimal amount, string currency)
    {
        if (amount < 0)
            throw new DomainException("O valor não pode ser negativo.");
        if (string.IsNullOrWhiteSpace(currency) || currency.Length != 3)
            throw new DomainException("A moeda deve ter 3 letras, como BRL.");

        Amount = amount;
        Currency = currency.ToUpperInvariant();
    }

    public Money Multiply(int quantity) => new(Amount * quantity, Currency);

    public Money Add(Money other)
    {
        if (other.Currency != Currency)
            throw new DomainException("Não é possível somar moedas diferentes.");
        return new Money(Amount + other.Amount, Currency);
    }
}
