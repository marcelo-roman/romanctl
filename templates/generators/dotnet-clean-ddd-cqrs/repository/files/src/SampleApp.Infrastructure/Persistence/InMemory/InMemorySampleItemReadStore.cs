using SampleApp.Application.SampleItems;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal sealed class InMemorySampleItemReadStore(InMemorySampleItemStore store) : ISampleItemReadStore
{
    public Task<SampleItemDto?> FindAsync(Guid id, CancellationToken cancellationToken) =>
        Task.FromResult(store.Items.TryGetValue(new SampleItemId(id), out var item) ? SampleItemDto.From(item) : null);

    public Task<IReadOnlyList<SampleItemDto>> ListAsync(int limit, CancellationToken cancellationToken)
    {
        IReadOnlyList<SampleItemDto> items = store.Items.Values
            .OrderByDescending(x => x.CreatedAt)
            .Take(limit)
            .Select(SampleItemDto.From)
            .ToList();
        return Task.FromResult(items);
    }
}
