using SampleApp.Api.Domain;
using SampleApp.Api.Domain.Orders;

namespace SampleApp.Tests.Orders;

public class OrderTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Total_soma_os_subtotais_dos_itens()
    {
        var order = Order.Create("Ana", Now);

        order.AddItem("Caneta", 2, new Money(1.50m, "BRL"));
        order.AddItem("Caderno", 1, new Money(10m, "BRL"));

        Assert.Equal(new Money(13m, "BRL"), order.Total);
    }

    [Fact]
    public void Pedido_novo_comeca_em_rascunho_com_total_zero_em_BRL()
    {
        var order = Order.Create(" Ana ", Now);

        Assert.Equal("Ana", order.CustomerName);
        Assert.Equal(OrderStatus.Draft, order.Status);
        Assert.Equal(new Money(0m, Order.DefaultCurrency), order.Total);
    }

    [Fact]
    public void Rejeita_cliente_sem_nome() =>
        Assert.Throws<DomainException>(() => Order.Create(" ", Now));

    [Fact]
    public void Nao_realiza_pedido_sem_itens()
    {
        var order = Order.Create("Ana", Now);

        Assert.Throws<DomainException>(() => order.Place());
    }

    [Fact]
    public void Realizar_o_pedido_impede_alteracoes()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem("Caneta", 2, new Money(1.50m, "BRL"));

        order.Place();

        Assert.Equal(OrderStatus.Placed, order.Status);
        Assert.Throws<DomainException>(() => order.AddItem("Lápis", 1, new Money(1m, "BRL")));
        Assert.Throws<DomainException>(() => order.Place());
    }

    [Fact]
    public void Nao_mistura_moedas_no_mesmo_pedido()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem("Caneta", 1, new Money(1m, "BRL"));

        Assert.Throws<DomainException>(() => order.AddItem("Pen", 1, new Money(1m, "USD")));
    }

    [Theory]
    [InlineData("", 1)]
    [InlineData("Caneta", 0)]
    public void Rejeita_item_invalido(string product, int quantity)
    {
        var order = Order.Create("Ana", Now);

        Assert.Throws<DomainException>(() => order.AddItem(product, quantity, new Money(1m, "BRL")));
    }

    [Theory]
    [InlineData(-1, "BRL")]
    [InlineData(1, "REAL")]
    [InlineData(1, "")]
    public void Rejeita_dinheiro_invalido(decimal amount, string currency) =>
        Assert.Throws<DomainException>(() => new Money(amount, currency));
}
