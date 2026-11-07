using {{.Name}}.Application.Orders;
using {{.Name}}.Domain.Orders;
using Microsoft.EntityFrameworkCore;

namespace {{.Name}}.Infrastructure.Persistence.EfCore;

internal sealed class OrderReadStore(AppDbContext db) : IOrderReadStore
{
    public async Task<OrderDto?> FindAsync(Guid id, CancellationToken cancellationToken)
    {
        var orderId = new OrderId(id);
        var order = await db.Orders.AsNoTracking().FirstOrDefaultAsync(o => o.Id == orderId, cancellationToken);
        return order is null ? null : OrderDto.From(order);
    }

    public async Task<IReadOnlyList<OrderDto>> ListAsync(int limit, CancellationToken cancellationToken)
    {
        var orders = await db.Orders.AsNoTracking()
            .OrderByDescending(order => order.CreatedAt)
            .Take(limit)
            .ToListAsync(cancellationToken);
        return orders.Select(OrderDto.From).ToList();
    }
}
