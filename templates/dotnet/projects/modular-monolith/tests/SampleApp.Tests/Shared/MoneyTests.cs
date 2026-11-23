using SampleApp.Shared.Domain;

namespace SampleApp.Tests.Shared;

public class MoneyTests
{
    [Fact]
    public void Normalizes_currency_to_uppercase() =>
        Assert.Equal(new Money(1m, "BRL"), new Money(1m, "brl"));

    [Fact]
    public void Adds_and_multiplies_in_the_same_currency()
    {
        var total = new Money(1.50m, "BRL").Multiply(2).Add(new Money(10m, "BRL"));

        Assert.Equal(new Money(13m, "BRL"), total);
    }

    [Fact]
    public void Does_not_add_different_currencies() =>
        Assert.Throws<DomainException>(() => new Money(1m, "BRL").Add(new Money(1m, "USD")));

    [Theory]
    [InlineData(-1, "BRL")]
    [InlineData(1, "REAL")]
    [InlineData(1, "")]
    public void Rejects_invalid_money(decimal amount, string currency) =>
        Assert.Throws<DomainException>(() => new Money(amount, currency));
}
