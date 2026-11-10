using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;

namespace SampleApp.Application.Orders.Queries;

public sealed record GetOrderEvents(Guid OrderId) : IQuery<IReadOnlyList<EventDto>>;

internal sealed class GetOrderEventsHandler(IEventStore store) : IQueryHandler<GetOrderEvents, IReadOnlyList<EventDto>>
{
    public async Task<IReadOnlyList<EventDto>> HandleAsync(GetOrderEvents query, CancellationToken cancellationToken)
    {
        var stream = await store.ReadStreamAsync(query.OrderId, cancellationToken);
        if (stream.Count == 0)
            throw new NotFoundException($"Pedido {query.OrderId} não encontrado.");

        return stream.Select(EventDto.From).ToList();
    }
}
