using SampleApp.Application.Common;
using SampleApp.Application.Messaging;
using SampleApp.Domain.Common;

namespace SampleApp.Infrastructure.Persistence.InMemory;

/// <summary>Guarda o que mudou no escopo atual e só grava em SaveChangesAsync, como um banco faria.</summary>
internal sealed class InMemoryUnitOfWork(IDomainEventPublisher publisher) : IUnitOfWork
{
    private readonly List<(IAggregateRoot Aggregate, Action Commit)> _tracked = [];

    public void Track(IAggregateRoot aggregate, Action commit) => _tracked.Add((aggregate, commit));

    public async Task SaveChangesAsync(CancellationToken cancellationToken)
    {
        foreach (var (_, commit) in _tracked)
            commit();

        var events = _tracked.SelectMany(item => item.Aggregate.DomainEvents).ToList();
        foreach (var (aggregate, _) in _tracked)
            aggregate.ClearDomainEvents();
        _tracked.Clear();

        foreach (var domainEvent in events)
            await publisher.PublishAsync(domainEvent, cancellationToken);
    }
}
