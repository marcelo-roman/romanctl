using SampleApp.Application.Messaging;

namespace SampleApp.Application.SampleParents.Queries;

public sealed record SampleItem(int Limit = 50) : IQuery<IReadOnlyList<SampleParentDto>>;

internal sealed class SampleItemHandler(ISampleParentReadStore store) : IQueryHandler<SampleItem, IReadOnlyList<SampleParentDto>>
{
    public Task<IReadOnlyList<SampleParentDto>> HandleAsync(SampleItem query, CancellationToken cancellationToken) =>
        store.ListAsync(Math.Clamp(query.Limit, 1, 200), cancellationToken);
}
