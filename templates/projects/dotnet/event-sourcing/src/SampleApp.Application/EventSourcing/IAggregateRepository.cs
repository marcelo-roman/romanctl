using SampleApp.Domain.Common;

namespace SampleApp.Application.EventSourcing;

public interface IAggregateRepository
{
    Task<TAggregate?> LoadAsync<TAggregate>(Guid id, CancellationToken cancellationToken)
        where TAggregate : EventSourcedAggregate, new();

    Task SaveAsync(EventSourcedAggregate aggregate, CancellationToken cancellationToken);
}
