using SampleApp.Application;
using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;
using SampleApp.Application.Orders.Commands;
using SampleApp.Application.Orders.Queries;
using SampleApp.Domain.Orders;
using SampleApp.Infrastructure.Persistence.InMemory;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Tests;

public class DispatcherTests
{
    private static IServiceProvider Build()
    {
        var services = new ServiceCollection().AddApplication().AddInMemoryPersistence();
        return services.BuildServiceProvider().CreateScope().ServiceProvider;
    }

    [Fact]
    public async Task Fluxo_completo_de_comandos_e_consultas()
    {
        var dispatcher = Build().GetRequiredService<IDispatcher>();

        var id = await dispatcher.SendAsync(new CreateOrder("Ana"));
        await dispatcher.SendAsync(new AddOrderItem(id, "Caneta", 2, 1.50m));
        await dispatcher.SendAsync(new PlaceOrder(id));

        var order = await dispatcher.QueryAsync(new GetOrderById(id));
        Assert.Equal("Placed", order.Status);
        Assert.Equal(3m, order.Total);
        Assert.Equal("BRL", order.Currency);
        Assert.Single(await dispatcher.QueryAsync(new ListOrders()));

        var events = await dispatcher.QueryAsync(new GetOrderEvents(id));
        Assert.Equal([1L, 2L, 3L], events.Select(e => e.Version));
        Assert.Equal(["OrderCreated", "ItemAdded", "OrderPlaced"], events.Select(e => e.Type));
    }

    [Fact]
    public async Task Leitura_vem_da_projecao_atualizada_pelos_eventos()
    {
        var provider = Build();
        var id = Guid.NewGuid();

        await provider.GetRequiredService<IEventStore>()
            .AppendAsync(id, 0, [new OrderCreated("Bia", DateTimeOffset.UtcNow)], CancellationToken.None);

        var order = await provider.GetRequiredService<IDispatcher>().QueryAsync(new GetOrderById(id));
        Assert.Equal("Bia", order.CustomerName);
        Assert.Equal("Draft", order.Status);
    }

    [Fact]
    public async Task Lista_os_mais_recentes_primeiro_e_respeita_o_limite()
    {
        var dispatcher = Build().GetRequiredService<IDispatcher>();
        await dispatcher.SendAsync(new CreateOrder("Ana"));
        await Task.Delay(5);
        var newest = await dispatcher.SendAsync(new CreateOrder("Bia"));

        var orders = await dispatcher.QueryAsync(new ListOrders(1));

        Assert.Equal(newest, Assert.Single(orders).Id);
    }

    [Fact]
    public async Task Pedido_inexistente_vira_NotFound()
    {
        var dispatcher = Build().GetRequiredService<IDispatcher>();

        await Assert.ThrowsAsync<NotFoundException>(() => dispatcher.QueryAsync(new GetOrderById(Guid.NewGuid())));
        await Assert.ThrowsAsync<NotFoundException>(() => dispatcher.QueryAsync(new GetOrderEvents(Guid.NewGuid())));
        await Assert.ThrowsAsync<NotFoundException>(() => dispatcher.SendAsync(new PlaceOrder(Guid.NewGuid())));
    }

    [Fact]
    public async Task Duas_alteracoes_sobre_a_mesma_versao_geram_conflito()
    {
        var provider = Build();
        var dispatcher = provider.GetRequiredService<IDispatcher>();
        var repository = provider.GetRequiredService<IAggregateRepository>();
        var id = await dispatcher.SendAsync(new CreateOrder("Ana"));

        var first = await repository.LoadAsync<Order>(id, CancellationToken.None);
        var second = await repository.LoadAsync<Order>(id, CancellationToken.None);
        first!.AddItem("Caneta", 1, new Money(1m, "BRL"));
        second!.AddItem("Lápis", 1, new Money(1m, "BRL"));
        await repository.SaveAsync(first, CancellationToken.None);

        await Assert.ThrowsAsync<ConcurrencyException>(() => repository.SaveAsync(second, CancellationToken.None));
        Assert.Single((await dispatcher.QueryAsync(new GetOrderById(id))).Items);
    }

    [Fact]
    public async Task Nao_inicia_um_fluxo_que_ja_existe()
    {
        var store = Build().GetRequiredService<IEventStore>();
        var id = Guid.NewGuid();
        await store.AppendAsync(id, 0, [new OrderCreated("Ana", DateTimeOffset.UtcNow)], CancellationToken.None);

        await Assert.ThrowsAsync<ConcurrencyException>(() =>
            store.AppendAsync(id, 0, [new OrderCreated("Bia", DateTimeOffset.UtcNow)], CancellationToken.None));
    }
}
