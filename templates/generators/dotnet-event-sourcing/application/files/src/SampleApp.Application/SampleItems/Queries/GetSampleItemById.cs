using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;

namespace SampleApp.Application.SampleItems.Queries;

public sealed record GetSampleItemById(Guid Id) : IQuery<SampleItemSummary>;

internal sealed class GetSampleItemByIdHandler(IReadModelStore<SampleItemSummary> store)
    : IQueryHandler<GetSampleItemById, SampleItemSummary>
{
    public async Task<SampleItemSummary> HandleAsync(GetSampleItemById query, CancellationToken cancellationToken) =>
        await store.FindAsync(query.Id, cancellationToken)
        ?? throw new NotFoundException($"SampleItem {query.Id} não encontrado.");
}
