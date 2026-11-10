using SampleApp.Domain.Common;
using SampleApp.Domain.Orders;

namespace SampleApp.Tests;

public class OrderTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);
    private static readonly OrderCreated Created = new("Ana", Now);
    private static readonly ItemAdded Pens = new("Caneta", 2, 1.50m, "BRL");

    private static AggregateScenario<Order> Scenario() => new();

    [Fact]
    public void Criar_gera_OrderCreated() =>
        Scenario()
            .WhenCreating(id => Order.Create(id, " Ana ", Now))
            .Then(new OrderCreated("Ana", Now));

    [Fact]
    public void Criar_exige_o_nome_do_cliente() =>
        Scenario()
            .WhenCreating(id => Order.Create(id, " ", Now))
            .ThenFails("O nome do cliente é obrigatório.");

    [Fact]
    public void Adicionar_item_gera_ItemAdded_com_moeda_padrao() =>
        Scenario()
            .Given(Created)
            .When(order => order.AddItem(" Caneta ", 2, new Money(1.50m, Order.DefaultCurrency)))
            .Then(Pens);

    [Theory]
    [InlineData("", 1, "O nome do produto é obrigatório.")]
    [InlineData("Caneta", 0, "A quantidade deve ser maior que zero.")]
    public void Rejeita_item_invalido(string product, int quantity, string message) =>
        Scenario()
            .Given(Created)
            .When(order => order.AddItem(product, quantity, new Money(1m, "BRL")))
            .ThenFails(message);

    [Fact]
    public void Nao_mistura_moedas_no_mesmo_pedido() =>
        Scenario()
            .Given(Created, Pens)
            .When(order => order.AddItem("Pen", 1, new Money(1m, "USD")))
            .ThenFails("Todos os itens do pedido devem usar a mesma moeda.");

    [Fact]
    public void Realizar_soma_os_itens() =>
        Scenario()
            .Given(Created, Pens, new ItemAdded("Caderno", 1, 10m, "BRL"))
            .When(order => order.Place(Now))
            .Then(new OrderPlaced(13m, "BRL", Now));

    [Fact]
    public void Nao_realiza_pedido_sem_itens() =>
        Scenario()
            .Given(Created)
            .When(order => order.Place(Now))
            .ThenFails("Um pedido sem itens não pode ser realizado.");

    [Fact]
    public void Pedido_realizado_nao_aceita_itens() =>
        Scenario()
            .Given(Created, Pens, new OrderPlaced(3m, "BRL", Now))
            .When(order => order.AddItem("Lápis", 1, new Money(1m, "BRL")))
            .ThenFails("Um pedido já realizado não pode ser alterado.");

    [Fact]
    public void Pedido_realizado_nao_e_realizado_de_novo() =>
        Scenario()
            .Given(Created, Pens, new OrderPlaced(3m, "BRL", Now))
            .When(order => order.Place(Now))
            .ThenFails("O pedido já foi realizado.");

    [Fact]
    public void Reconstroi_o_estado_e_a_versao_a_partir_dos_eventos()
    {
        var order = new Order();

        order.LoadFromHistory(Guid.NewGuid(), [Created, Pens, new OrderPlaced(3m, "BRL", Now)]);

        Assert.Equal(3, order.Version);
        Assert.Equal(OrderStatus.Placed, order.Status);
        Assert.Equal(new Money(3m, "BRL"), order.Total);
        Assert.Empty(order.PendingEvents);
    }

    [Fact]
    public void Confirmar_os_eventos_avanca_a_versao()
    {
        var order = Order.Create(Guid.NewGuid(), "Ana", Now);
        order.AddItem("Caneta", 1, new Money(1m, "BRL"));

        order.MarkCommitted();

        Assert.Equal(2, order.Version);
        Assert.Empty(order.PendingEvents);
    }

    [Theory]
    [InlineData(-1, "BRL", "O valor não pode ser negativo.")]
    [InlineData(1, "REAL", "A moeda deve ter 3 letras, como BRL.")]
    public void Rejeita_dinheiro_invalido(decimal amount, string currency, string message) =>
        Assert.Equal(message, Assert.Throws<DomainException>(() => new Money(amount, currency)).Message);
}
