using {{.Name}}.Domain.Orders;
using Microsoft.EntityFrameworkCore;

namespace {{.Name}}.Infrastructure.Persistence.EfCore;

internal sealed class OrderRepository(AppDbContext db) : IOrderRepository
{
    public void Add(Order order) => db.Orders.Add(order);

    public Task<Order?> GetByIdAsync(OrderId id, CancellationToken cancellationToken) =>
        db.Orders.FirstOrDefaultAsync(order => order.Id == id, cancellationToken);
}
