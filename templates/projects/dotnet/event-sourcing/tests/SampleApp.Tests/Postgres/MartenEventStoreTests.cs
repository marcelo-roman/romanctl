using SampleApp.Application;
using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;
using SampleApp.Application.Orders;
using SampleApp.Application.Orders.Commands;
using SampleApp.Application.Orders.Queries;
using SampleApp.Domain.Orders;
using SampleApp.Infrastructure;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Tests.Postgres;

[Collection(PostgresCollection.Name)]
public class MartenEventStoreTests
{
    private static ServiceProvider Build()
    {
        var configuration = new ConfigurationBuilder().AddEnvironmentVariables().Build();
        return new ServiceCollection().AddLogging().AddApplication().AddInfrastructure(configuration).BuildServiceProvider();
    }

    private static async Task<T> InScope<T>(ServiceProvider provider, Func<IServiceProvider, Task<T>> work)
    {
        await using var scope = provider.CreateAsyncScope();
        return await work(scope.ServiceProvider);
    }

    [PostgresFact]
    public async Task Grava_le_o_fluxo_e_atualiza_a_projecao()
    {
        await using var provider = Build();

        var id = await InScope(provider, services => services.GetRequiredService<IDispatcher>().SendAsync(new CreateOrder("Ana")));
        await InScope(provider, services => services.GetRequiredService<IDispatcher>().SendAsync(new AddOrderItem(id, "Caneta", 2, 1.50m)));
        await InScope(provider, services => services.GetRequiredService<IDispatcher>().SendAsync(new PlaceOrder(id)));

        var stream = await InScope(provider, services => services.GetRequiredService<IEventStore>().ReadStreamAsync(id, CancellationToken.None));
        Assert.Equal([1L, 2L, 3L], stream.Select(stored => stored.Version));
        Assert.Equal(new ItemAdded("Caneta", 2, 1.50m, "BRL"), stream[1].Data);

        var summary = await InScope(provider, services => services.GetRequiredService<IReadModelStore<OrderSummary>>().FindAsync(id, CancellationToken.None));
        Assert.NotNull(summary);
        Assert.Equal("Placed", summary.Status);
        Assert.Equal(3m, summary.Total);
        Assert.Single(summary.Items);

        var recent = await InScope(provider, services => services.GetRequiredService<IDispatcher>().QueryAsync(new ListOrders(200)));
        Assert.Contains(recent, order => order.Id == id);
        var events = await InScope(provider, services => services.GetRequiredService<IDispatcher>().QueryAsync(new GetOrderEvents(id)));
        Assert.Equal(["OrderCreated", "ItemAdded", "OrderPlaced"], events.Select(e => e.Type));
    }

    [PostgresFact]
    public async Task Versao_desatualizada_gera_conflito()
    {
        await using var provider = Build();
        var id = await InScope(provider, services => services.GetRequiredService<IDispatcher>().SendAsync(new CreateOrder("Ana")));

        var stale = await InScope(provider, services => services.GetRequiredService<IAggregateRepository>().LoadAsync<Order>(id, CancellationToken.None));
        await InScope(provider, services => services.GetRequiredService<IDispatcher>().SendAsync(new AddOrderItem(id, "Caneta", 1, 1m)));
        stale!.AddItem("Lápis", 1, new Money(1m, "BRL"));

        await Assert.ThrowsAsync<ConcurrencyException>(() =>
            InScope(provider, async services =>
            {
                await services.GetRequiredService<IAggregateRepository>().SaveAsync(stale, CancellationToken.None);
                return true;
            }));
    }

    [PostgresFact]
    public async Task Nao_inicia_um_fluxo_que_ja_existe()
    {
        await using var provider = Build();
        var id = await InScope(provider, services => services.GetRequiredService<IDispatcher>().SendAsync(new CreateOrder("Ana")));

        await Assert.ThrowsAsync<ConcurrencyException>(() =>
            InScope(provider, async services =>
            {
                await services.GetRequiredService<IEventStore>().AppendAsync(id, 0, [new OrderCreated("Bia", DateTimeOffset.UtcNow)], CancellationToken.None);
                return true;
            }));
    }
}
