using SampleApp.Ordering.Domain;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Ordering.Infrastructure;

internal sealed class OrderRepository(OrderingDbContext db) : IOrderRepository
{
    public void Add(Order order) => db.Set<Order>().Add(order);

    public Task<Order?> FindAsync(Guid id, CancellationToken cancellationToken) =>
        db.Set<Order>().FirstOrDefaultAsync(order => order.Id == id, cancellationToken);

    public async Task<IReadOnlyList<Order>> ListRecentAsync(int limit, CancellationToken cancellationToken) =>
        await db.Set<Order>().AsNoTracking()
            .OrderByDescending(order => order.CreatedAt)
            .Take(limit)
            .ToListAsync(cancellationToken);

    public Task SaveChangesAsync(CancellationToken cancellationToken) => db.SaveChangesAsync(cancellationToken);
}
