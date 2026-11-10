using SampleApp.Domain.Common;

namespace SampleApp.Application.EventSourcing;

internal sealed class AggregateRepository(IEventStore store) : IAggregateRepository
{
    public async Task<TAggregate?> LoadAsync<TAggregate>(Guid id, CancellationToken cancellationToken)
        where TAggregate : EventSourcedAggregate, new()
    {
        var history = await store.ReadStreamAsync(id, cancellationToken);
        if (history.Count == 0)
            return null;

        var aggregate = new TAggregate();
        aggregate.LoadFromHistory(id, history.Select(stored => stored.Data));
        return aggregate;
    }

    public async Task SaveAsync(EventSourcedAggregate aggregate, CancellationToken cancellationToken)
    {
        if (aggregate.PendingEvents.Count == 0)
            return;

        await store.AppendAsync(aggregate.Id, aggregate.Version, aggregate.PendingEvents, cancellationToken);
        aggregate.MarkCommitted();
    }
}
