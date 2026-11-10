namespace SampleApp.Application.Orders;

/// <summary>Lado de leitura do CQRS: devolve DTOs e não passa pelo agregado para alterar nada.</summary>
public interface IOrderReadStore
{
    Task<OrderDto?> FindAsync(Guid id, CancellationToken cancellationToken);
    Task<IReadOnlyList<OrderDto>> ListAsync(int limit, CancellationToken cancellationToken);
}
