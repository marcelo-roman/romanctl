using SampleApp.Application;
using SampleApp.Application.Common;
using SampleApp.Application.Messaging;
using SampleApp.Application.Orders;
using SampleApp.Application.Orders.Commands;
using SampleApp.Application.Orders.Queries;
using SampleApp.Domain.Common;
using SampleApp.Domain.Orders;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Tests;

public class DispatcherTests
{
    private sealed class FakeStore(IDomainEventPublisher publisher) : IOrderRepository, IOrderReadStore, IUnitOfWork
    {
        private readonly Dictionary<OrderId, Order> _orders = [];

        public void Add(Order order) => _orders[order.Id] = order;

        public Task<Order?> GetByIdAsync(OrderId id, CancellationToken cancellationToken) =>
            Task.FromResult(_orders.GetValueOrDefault(id));

        public Task<OrderDto?> FindAsync(Guid id, CancellationToken cancellationToken) =>
            Task.FromResult(_orders.TryGetValue(new OrderId(id), out var order) ? OrderDto.From(order) : null);

        public Task<IReadOnlyList<OrderDto>> ListAsync(int limit, CancellationToken cancellationToken) =>
            Task.FromResult<IReadOnlyList<OrderDto>>(_orders.Values.Take(limit).Select(OrderDto.From).ToList());

        public async Task SaveChangesAsync(CancellationToken cancellationToken)
        {
            foreach (var order in _orders.Values)
            {
                var events = order.DomainEvents.ToList();
                order.ClearDomainEvents();
                foreach (var domainEvent in events)
                    await publisher.PublishAsync(domainEvent, cancellationToken);
            }
        }
    }

    private sealed class RecordingHandler(List<IDomainEvent> received) : IDomainEventHandler<OrderPlaced>
    {
        public Task HandleAsync(OrderPlaced domainEvent, CancellationToken cancellationToken)
        {
            received.Add(domainEvent);
            return Task.CompletedTask;
        }
    }

    private static (IDispatcher Dispatcher, List<IDomainEvent> Events) Build()
    {
        var events = new List<IDomainEvent>();
        var services = new ServiceCollection().AddLogging().AddApplication();
        services.AddScoped<FakeStore>();
        services.AddScoped<IOrderRepository>(provider => provider.GetRequiredService<FakeStore>());
        services.AddScoped<IOrderReadStore>(provider => provider.GetRequiredService<FakeStore>());
        services.AddScoped<IUnitOfWork>(provider => provider.GetRequiredService<FakeStore>());
        services.AddScoped<IDomainEventHandler<OrderPlaced>>(_ => new RecordingHandler(events));

        var scope = services.BuildServiceProvider().CreateScope();
        return (scope.ServiceProvider.GetRequiredService<IDispatcher>(), events);
    }

    [Fact]
    public async Task Fluxo_completo_de_comandos_e_consultas()
    {
        var (dispatcher, events) = Build();

        var id = await dispatcher.SendAsync(new CreateOrder("Ana"));
        await dispatcher.SendAsync(new AddOrderItem(id, "Caneta", 2, 1.50m));
        await dispatcher.SendAsync(new PlaceOrder(id));

        var order = await dispatcher.QueryAsync(new GetOrderById(id));
        Assert.Equal("Placed", order.Status);
        Assert.Equal(3m, order.Total);
        Assert.Single(await dispatcher.QueryAsync(new ListOrders()));
        Assert.IsType<OrderPlaced>(Assert.Single(events));
    }

    [Fact]
    public async Task Pedido_inexistente_vira_NotFound()
    {
        var (dispatcher, _) = Build();

        await Assert.ThrowsAsync<NotFoundException>(() => dispatcher.QueryAsync(new GetOrderById(Guid.NewGuid())));
        await Assert.ThrowsAsync<NotFoundException>(() => dispatcher.SendAsync(new PlaceOrder(Guid.NewGuid())));
    }
}
