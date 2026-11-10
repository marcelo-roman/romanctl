using SampleApp.Domain.Common;
using SampleApp.Domain.Orders;

namespace SampleApp.Tests;

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
    public void Nao_realiza_pedido_sem_itens()
    {
        var order = Order.Create("Ana", Now);

        Assert.Throws<DomainException>(() => order.Place(Now));
    }

    [Fact]
    public void Realizar_o_pedido_levanta_evento_e_impede_alteracoes()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem("Caneta", 2, new Money(1.50m, "BRL"));

        order.Place(Now);

        var placed = Assert.IsType<OrderPlaced>(Assert.Single(order.DomainEvents));
        Assert.Equal(3m, placed.Total);
        Assert.Throws<DomainException>(() => order.AddItem("Lápis", 1, new Money(1m, "BRL")));
        Assert.Throws<DomainException>(() => order.Place(Now));
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

    [Fact]
    public void Rejeita_valor_negativo() =>
        Assert.Throws<DomainException>(() => new Money(-1m, "BRL"));
}
