using Marten;
using Marten.Exceptions;
using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Domain.Common;

namespace SampleApp.Infrastructure.Persistence.Postgres;

internal sealed class MartenEventStore(IDocumentSession session) : IEventStore
{
    public async Task<IReadOnlyList<StoredEvent>> ReadStreamAsync(Guid streamId, CancellationToken cancellationToken)
    {
        var events = await session.Events.FetchStreamAsync(streamId, token: cancellationToken);
        return events
            .Select(stored => new StoredEvent(streamId, stored.Version, (IDomainEvent)stored.Data, stored.Timestamp))
            .ToList();
    }

    public async Task AppendAsync(Guid streamId, long expectedVersion, IReadOnlyList<IDomainEvent> events, CancellationToken cancellationToken)
    {
        if (expectedVersion == 0)
            session.Events.StartStream(streamId, events);
        else
            session.Events.Append(streamId, expectedVersion + events.Count, events);

        try
        {
            await session.SaveChangesAsync(cancellationToken);
        }
        catch (Exception exception) when (exception is JasperFx.ConcurrencyException or ExistingStreamIdCollisionException)
        {
            throw new ConcurrencyException(streamId, exception);
        }
    }
}
