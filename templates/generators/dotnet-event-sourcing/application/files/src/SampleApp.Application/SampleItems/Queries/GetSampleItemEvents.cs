using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;

namespace SampleApp.Application.SampleItems.Queries;

public sealed record GetSampleItemEvents(Guid Id) : IQuery<IReadOnlyList<EventDto>>;

internal sealed class GetSampleItemEventsHandler(IEventStore store)
    : IQueryHandler<GetSampleItemEvents, IReadOnlyList<EventDto>>
{
    public async Task<IReadOnlyList<EventDto>> HandleAsync(GetSampleItemEvents query, CancellationToken cancellationToken)
    {
        var stream = await store.ReadStreamAsync(query.Id, cancellationToken);
        if (stream.Count == 0)
            throw new NotFoundException($"SampleItem {query.Id} não encontrado.");

        return stream.Select(EventDto.From).ToList();
    }
}
