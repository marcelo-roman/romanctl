using SampleApp.Application.Messaging;

namespace SampleApp.Application.SampleItems.Queries;

public sealed record ListSampleItems(int Limit = 50) : IQuery<IReadOnlyList<SampleItemDto>>;

internal sealed class ListSampleItemsHandler(ISampleItemReadStore store) : IQueryHandler<ListSampleItems, IReadOnlyList<SampleItemDto>>
{
    public Task<IReadOnlyList<SampleItemDto>> HandleAsync(ListSampleItems query, CancellationToken cancellationToken) =>
        store.ListAsync(Math.Clamp(query.Limit, 1, 200), cancellationToken);
}
