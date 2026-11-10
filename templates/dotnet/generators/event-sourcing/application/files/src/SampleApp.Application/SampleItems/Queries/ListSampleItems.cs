using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;

namespace SampleApp.Application.SampleItems.Queries;

public sealed record ListSampleItems(int Limit = 50) : IQuery<IReadOnlyList<SampleItemSummary>>;

internal sealed class ListSampleItemsHandler(IReadModelStore<SampleItemSummary> store)
    : IQueryHandler<ListSampleItems, IReadOnlyList<SampleItemSummary>>
{
    public Task<IReadOnlyList<SampleItemSummary>> HandleAsync(ListSampleItems query, CancellationToken cancellationToken) =>
        store.ListAsync(Math.Clamp(query.Limit, 1, 200), cancellationToken);
}
