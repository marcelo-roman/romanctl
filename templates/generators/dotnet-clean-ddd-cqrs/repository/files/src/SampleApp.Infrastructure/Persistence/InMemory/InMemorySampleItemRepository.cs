using SampleApp.Domain.SampleItems;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal sealed class InMemorySampleItemRepository(InMemorySampleItemStore store, InMemoryUnitOfWork unitOfWork) : ISampleItemRepository
{
    public void Add(SampleItem sampleItem) =>
        unitOfWork.Track(sampleItem, () => store.Items[sampleItem.Id] = sampleItem);

    public void Remove(SampleItem sampleItem) =>
        unitOfWork.Track(sampleItem, () => store.Items.TryRemove(sampleItem.Id, out _));

    public Task<SampleItem?> GetByIdAsync(SampleItemId id, CancellationToken cancellationToken)
    {
        if (store.Items.TryGetValue(id, out var item))
            unitOfWork.Track(item, () => store.Items[id] = item);
        return Task.FromResult(item);
    }
}
