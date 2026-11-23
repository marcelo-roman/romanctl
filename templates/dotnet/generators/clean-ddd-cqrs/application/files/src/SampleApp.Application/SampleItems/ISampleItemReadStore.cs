namespace SampleApp.Application.SampleItems;

/// <summary>CQRS read side: returns DTOs and never goes through the aggregate to change anything.</summary>
public interface ISampleItemReadStore
{
    Task<SampleItemDto?> FindAsync(Guid id, CancellationToken cancellationToken);
    Task<IReadOnlyList<SampleItemDto>> ListAsync(int limit, CancellationToken cancellationToken);
}
