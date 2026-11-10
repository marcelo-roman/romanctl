namespace SampleApp.Domain.SampleItems;

public interface ISampleItemRepository
{
    void Add(SampleItem sampleItem);
    void Remove(SampleItem sampleItem);
    Task<SampleItem?> GetByIdAsync(SampleItemId id, CancellationToken cancellationToken);
}
