namespace SampleApp.Ordering.Domain;

internal interface IOrderRepository
{
    void Add(Order order);

    Task<Order?> FindAsync(Guid id, CancellationToken cancellationToken);

    Task<IReadOnlyList<Order>> ListRecentAsync(int limit, CancellationToken cancellationToken);

    Task SaveChangesAsync(CancellationToken cancellationToken);
}
