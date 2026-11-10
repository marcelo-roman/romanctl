using SampleApp.Application.Orders;
using SampleApp.Domain.Orders;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal sealed class InMemoryOrderReadStore(InMemoryDatabase db) : IOrderReadStore
{
    public Task<OrderDto?> FindAsync(Guid id, CancellationToken cancellationToken) =>
        Task.FromResult(db.Orders.TryGetValue(new OrderId(id), out var order) ? OrderDto.From(order) : null);

    public Task<IReadOnlyList<OrderDto>> ListAsync(int limit, CancellationToken cancellationToken)
    {
        IReadOnlyList<OrderDto> orders = db.Orders.Values
            .OrderByDescending(order => order.CreatedAt)
            .Take(limit)
            .Select(OrderDto.From)
            .ToList();
        return Task.FromResult(orders);
    }
}
