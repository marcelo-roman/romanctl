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
    public async Task Full_flow_of_commands_and_queries()
    {
        var dispatcher = Build().GetRequiredService<IDispatcher>();

        var id = await dispatcher.SendAsync(new CreateOrder("Ana"));
        await dispatcher.SendAsync(new AddOrderItem(id, "Pen", 2, 1.50m));
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
    public async Task Read_comes_from_the_projection_updated_by_events()
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
    public async Task Lists_newest_first_and_respects_the_limit()
    {
        var dispatcher = Build().GetRequiredService<IDispatcher>();
        await dispatcher.SendAsync(new CreateOrder("Ana"));
        await Task.Delay(5);
        var newest = await dispatcher.SendAsync(new CreateOrder("Bia"));

        var orders = await dispatcher.QueryAsync(new ListOrders(1));

        Assert.Equal(newest, Assert.Single(orders).Id);
    }

    [Fact]
    public async Task Unknown_order_becomes_NotFound()
    {
        var dispatcher = Build().GetRequiredService<IDispatcher>();

        await Assert.ThrowsAsync<NotFoundException>(() => dispatcher.QueryAsync(new GetOrderById(Guid.NewGuid())));
        await Assert.ThrowsAsync<NotFoundException>(() => dispatcher.QueryAsync(new GetOrderEvents(Guid.NewGuid())));
        await Assert.ThrowsAsync<NotFoundException>(() => dispatcher.SendAsync(new PlaceOrder(Guid.NewGuid())));
    }

    [Fact]
    public async Task Two_changes_on_the_same_version_cause_conflict()
    {
        var provider = Build();
        var dispatcher = provider.GetRequiredService<IDispatcher>();
        var repository = provider.GetRequiredService<IAggregateRepository>();
        var id = await dispatcher.SendAsync(new CreateOrder("Ana"));

        var first = await repository.LoadAsync<Order>(id, CancellationToken.None);
        var second = await repository.LoadAsync<Order>(id, CancellationToken.None);
        first!.AddItem("Pen", 1, new Money(1m, "BRL"));
        second!.AddItem("Pencil", 1, new Money(1m, "BRL"));
        await repository.SaveAsync(first, CancellationToken.None);

        await Assert.ThrowsAsync<ConcurrencyException>(() => repository.SaveAsync(second, CancellationToken.None));
        Assert.Single((await dispatcher.QueryAsync(new GetOrderById(id))).Items);
    }

    [Fact]
    public async Task Does_not_start_a_stream_that_already_exists()
    {
        var store = Build().GetRequiredService<IEventStore>();
        var id = Guid.NewGuid();
        await store.AppendAsync(id, 0, [new OrderCreated("Ana", DateTimeOffset.UtcNow)], CancellationToken.None);

        await Assert.ThrowsAsync<ConcurrencyException>(() =>
            store.AppendAsync(id, 0, [new OrderCreated("Bia", DateTimeOffset.UtcNow)], CancellationToken.None));
    }
}
