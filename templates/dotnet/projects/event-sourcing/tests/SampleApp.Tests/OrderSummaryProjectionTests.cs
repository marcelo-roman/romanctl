using SampleApp.Application.Orders;
using SampleApp.Domain.Common;
using SampleApp.Domain.Orders;

namespace SampleApp.Tests;

public class OrderSummaryProjectionTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);
    private static readonly Guid Id = Guid.NewGuid();

    private static OrderSummary? Project(params IDomainEvent[] events)
    {
        var projection = new OrderSummaryProjection();
        return events.Aggregate((OrderSummary?)null, (current, domainEvent) => projection.Evolve(current, Id, domainEvent));
    }

    [Fact]
    public void Monta_o_resumo_a_partir_dos_eventos()
    {
        var summary = Project(
            new OrderCreated("Ana", Now),
            new ItemAdded("Caneta", 2, 1.50m, "BRL"),
            new ItemAdded("Caderno", 1, 10m, "BRL"),
            new OrderPlaced(13m, "BRL", Now.AddMinutes(5)));

        Assert.NotNull(summary);
        Assert.Equal(Id, summary.Id);
        Assert.Equal("Ana", summary.CustomerName);
        Assert.Equal("Placed", summary.Status);
        Assert.Equal(13m, summary.Total);
        Assert.Equal("BRL", summary.Currency);
        Assert.Equal(Now, summary.CreatedAt);
        Assert.Equal(Now.AddMinutes(5), summary.PlacedAt);
        Assert.Equal([new OrderSummaryItem("Caneta", 2, 1.50m, 3m), new OrderSummaryItem("Caderno", 1, 10m, 10m)], summary.Items);
    }

    [Fact]
    public void Pedido_novo_fica_em_rascunho_com_moeda_padrao()
    {
        var summary = Project(new OrderCreated("Ana", Now));

        Assert.NotNull(summary);
        Assert.Equal("Draft", summary.Status);
        Assert.Equal(Order.DefaultCurrency, summary.Currency);
        Assert.Empty(summary.Items);
    }

    [Fact]
    public void Ignora_eventos_sem_a_criacao() =>
        Assert.Null(Project(new ItemAdded("Caneta", 1, 1m, "BRL")));

    [Fact]
    public void Declara_os_eventos_que_trata() =>
        Assert.Equal(
            [typeof(OrderCreated), typeof(ItemAdded), typeof(OrderPlaced)],
            new OrderSummaryProjection().EventTypes);
}
