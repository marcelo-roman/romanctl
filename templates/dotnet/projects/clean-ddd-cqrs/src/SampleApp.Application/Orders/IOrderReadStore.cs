namespace SampleApp.Application.Orders;

/// <summary>CQRS read side: returns DTOs and never goes through the aggregate to change anything.</summary>
public interface IOrderReadStore
{
    Task<OrderDto?> FindAsync(Guid id, CancellationToken cancellationToken);
    Task<IReadOnlyList<OrderDto>> ListAsync(int limit, CancellationToken cancellationToken);
}
