using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;
using {{.Name}}.Domain.Orders;

namespace {{.Name}}.Infrastructure.Persistence.InMemory;

internal sealed class InMemoryOrderRepository(InMemoryDatabase db, IDomainEventPublisher publisher)
    : IOrderRepository, IUnitOfWork
{
    private readonly Dictionary<OrderId, Order> _tracked = [];

    public void Add(Order order) => _tracked[order.Id] = order;

    public Task<Order?> GetByIdAsync(OrderId id, CancellationToken cancellationToken)
    {
        if (db.Orders.TryGetValue(id, out var order))
            _tracked[id] = order;
        return Task.FromResult(order);
    }

    public async Task SaveChangesAsync(CancellationToken cancellationToken)
    {
        foreach (var (id, order) in _tracked)
            db.Orders[id] = order;

        var events = _tracked.Values.SelectMany(order => order.DomainEvents).ToList();
        foreach (var order in _tracked.Values)
            order.ClearDomainEvents();
        foreach (var domainEvent in events)
            await publisher.PublishAsync(domainEvent, cancellationToken);
    }
}
