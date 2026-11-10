namespace SampleApp.SampleParent.Domain;

internal interface ISampleItemRepository
{
    void Add(SampleItem sampleItem);

    void Remove(SampleItem sampleItem);

    Task<SampleItem?> FindAsync(Guid id, CancellationToken cancellationToken);

    Task<IReadOnlyList<SampleItem>> ListRecentAsync(int limit, CancellationToken cancellationToken);

    Task SaveChangesAsync(CancellationToken cancellationToken);
}
