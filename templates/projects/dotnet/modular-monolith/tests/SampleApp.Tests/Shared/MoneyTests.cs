using SampleApp.Shared.Domain;

namespace SampleApp.Tests.Shared;

public class MoneyTests
{
    [Fact]
    public void Normaliza_a_moeda_em_maiusculas() =>
        Assert.Equal(new Money(1m, "BRL"), new Money(1m, "brl"));

    [Fact]
    public void Soma_e_multiplica_na_mesma_moeda()
    {
        var total = new Money(1.50m, "BRL").Multiply(2).Add(new Money(10m, "BRL"));

        Assert.Equal(new Money(13m, "BRL"), total);
    }

    [Fact]
    public void Nao_soma_moedas_diferentes() =>
        Assert.Throws<DomainException>(() => new Money(1m, "BRL").Add(new Money(1m, "USD")));

    [Theory]
    [InlineData(-1, "BRL")]
    [InlineData(1, "REAL")]
    [InlineData(1, "")]
    public void Rejeita_dinheiro_invalido(decimal amount, string currency) =>
        Assert.Throws<DomainException>(() => new Money(amount, currency));
}
