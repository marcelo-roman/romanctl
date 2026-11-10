namespace SampleApp.Domain.Orders;

public interface IOrderRepository
{
    void Add(Order order);
    Task<Order?> GetByIdAsync(OrderId id, CancellationToken cancellationToken);
}
