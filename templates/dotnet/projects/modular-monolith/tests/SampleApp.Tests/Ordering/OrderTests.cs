using SampleApp.Ordering.Domain;
using SampleApp.Shared.Domain;

namespace SampleApp.Tests.Ordering;

public class OrderTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);
    private static readonly Guid Pen = Guid.NewGuid();
    private static readonly Guid Notebook = Guid.NewGuid();

    [Fact]
    public void Total_soma_os_subtotais_dos_itens()
    {
        var order = Order.Create("Ana", Now);

        order.AddItem(Pen, "Caneta", 2, new Money(1.50m, "BRL"));
        order.AddItem(Notebook, "Caderno", 1, new Money(10m, "BRL"));

        Assert.Equal(new Money(13m, "BRL"), order.Total);
    }

    [Fact]
    public void Pedido_novo_comeca_em_rascunho_com_total_zero_em_BRL()
    {
        var order = Order.Create(" Ana ", Now);

        Assert.Equal("Ana", order.CustomerName);
        Assert.Equal(OrderStatus.Draft, order.Status);
        Assert.Equal(Money.Zero(Money.DefaultCurrency), order.Total);
        Assert.Null(order.PlacedAt);
    }

    [Fact]
    public void Rejeita_cliente_sem_nome() =>
        Assert.Throws<DomainException>(() => Order.Create(" ", Now));

    [Fact]
    public void Nao_realiza_pedido_sem_itens()
    {
        var order = Order.Create("Ana", Now);

        Assert.Throws<DomainException>(() => order.Place(Now));
    }

    [Fact]
    public void Realizar_o_pedido_impede_alteracoes()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem(Pen, "Caneta", 2, new Money(1.50m, "BRL"));

        order.Place(Now.AddMinutes(1));

        Assert.Equal(OrderStatus.Placed, order.Status);
        Assert.Equal(Now.AddMinutes(1), order.PlacedAt);
        Assert.Throws<DomainException>(() => order.AddItem(Notebook, "Caderno", 1, new Money(1m, "BRL")));
        Assert.Throws<DomainException>(() => order.Place(Now));
    }

    [Fact]
    public void Nao_mistura_moedas_no_mesmo_pedido()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem(Pen, "Caneta", 1, new Money(1m, "BRL"));

        Assert.Throws<DomainException>(() => order.AddItem(Notebook, "Notebook", 1, new Money(1m, "USD")));
    }

    [Theory]
    [InlineData(true, "", 1)]
    [InlineData(true, "Caneta", 0)]
    [InlineData(false, "Caneta", 1)]
    public void Rejeita_item_invalido(bool withProduct, string productName, int quantity)
    {
        var order = Order.Create("Ana", Now);
        var productId = withProduct ? Pen : Guid.Empty;

        Assert.Throws<DomainException>(() => order.AddItem(productId, productName, quantity, new Money(1m, "BRL")));
    }
}
