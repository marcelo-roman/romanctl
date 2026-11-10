using SampleApp.Domain.Common;

namespace SampleApp.Application.EventSourcing;

public interface IEventStore
{
    Task<IReadOnlyList<StoredEvent>> ReadStreamAsync(Guid streamId, CancellationToken cancellationToken);

    Task AppendAsync(Guid streamId, long expectedVersion, IReadOnlyList<IDomainEvent> events, CancellationToken cancellationToken);
}
