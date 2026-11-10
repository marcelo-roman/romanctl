using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Domain.Common;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal sealed class InMemoryEventStore(IEnumerable<IInMemoryProjection> projections, TimeProvider clock) : IEventStore
{
    private readonly Dictionary<Guid, List<StoredEvent>> _streams = [];
    private readonly Lock _gate = new();

    public Task<IReadOnlyList<StoredEvent>> ReadStreamAsync(Guid streamId, CancellationToken cancellationToken)
    {
        lock (_gate)
        {
            IReadOnlyList<StoredEvent> stream = _streams.TryGetValue(streamId, out var events) ? events.ToList() : [];
            return Task.FromResult(stream);
        }
    }

    public Task AppendAsync(Guid streamId, long expectedVersion, IReadOnlyList<IDomainEvent> events, CancellationToken cancellationToken)
    {
        lock (_gate)
        {
            var stream = _streams.GetValueOrDefault(streamId) ?? [];
            if (stream.Count != expectedVersion)
                throw new ConcurrencyException(streamId);

            var recordedAt = clock.GetUtcNow();
            var appended = events
                .Select((domainEvent, index) => new StoredEvent(streamId, expectedVersion + index + 1, domainEvent, recordedAt))
                .ToList();

            stream.AddRange(appended);
            _streams[streamId] = stream;
            foreach (var stored in appended)
            foreach (var projection in projections)
                projection.Apply(stored);
        }

        return Task.CompletedTask;
    }
}
