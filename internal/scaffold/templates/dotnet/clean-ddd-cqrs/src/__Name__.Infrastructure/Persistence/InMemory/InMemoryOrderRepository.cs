using {{.Name}}.Domain.Orders;

namespace {{.Name}}.Infrastructure.Persistence.InMemory;

internal sealed class InMemoryOrderRepository(InMemoryDatabase db, InMemoryUnitOfWork unitOfWork) : IOrderRepository
{
    public void Add(Order order) => unitOfWork.Track(order, () => db.Orders[order.Id] = order);

    public Task<Order?> GetByIdAsync(OrderId id, CancellationToken cancellationToken)
    {
        if (db.Orders.TryGetValue(id, out var order))
            unitOfWork.Track(order, () => db.Orders[id] = order);
        return Task.FromResult(order);
    }
}
