using SampleApp.Application.Common;
using SampleApp.Application.Messaging;

namespace SampleApp.Application.SampleItems.Queries;

public sealed record GetSampleItemById(Guid Id) : IQuery<SampleItemDto>;

internal sealed class GetSampleItemByIdHandler(ISampleItemReadStore store) : IQueryHandler<GetSampleItemById, SampleItemDto>
{
    public async Task<SampleItemDto> HandleAsync(GetSampleItemById query, CancellationToken cancellationToken) =>
        await store.FindAsync(query.Id, cancellationToken)
        ?? throw new NotFoundException($"SampleItem {query.Id} não encontrado.");
}
