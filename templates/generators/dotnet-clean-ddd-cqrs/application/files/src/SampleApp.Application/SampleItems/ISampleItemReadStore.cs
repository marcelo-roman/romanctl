namespace SampleApp.Application.SampleItems;

/// <summary>Lado de leitura do CQRS: devolve DTOs e não passa pelo agregado para alterar nada.</summary>
public interface ISampleItemReadStore
{
    Task<SampleItemDto?> FindAsync(Guid id, CancellationToken cancellationToken);
    Task<IReadOnlyList<SampleItemDto>> ListAsync(int limit, CancellationToken cancellationToken);
}
