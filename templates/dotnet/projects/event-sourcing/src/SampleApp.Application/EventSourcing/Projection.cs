using SampleApp.Domain.Common;

namespace SampleApp.Application.EventSourcing;

public abstract class Projection<TView> where TView : class, IReadModel
{
    private readonly Dictionary<Type, Func<TView?, Guid, IDomainEvent, TView?>> _handlers = [];

    public IReadOnlyCollection<Type> EventTypes => _handlers.Keys;

    public TView? Evolve(TView? current, Guid streamId, IDomainEvent domainEvent) =>
        _handlers.TryGetValue(domainEvent.GetType(), out var handle) ? handle(current, streamId, domainEvent) : current;

    protected void Creates<TEvent>(Func<Guid, TEvent, TView> create) where TEvent : IDomainEvent =>
        _handlers[typeof(TEvent)] = (_, streamId, domainEvent) => create(streamId, (TEvent)domainEvent);

    protected void Updates<TEvent>(Func<TView, TEvent, TView> update) where TEvent : IDomainEvent =>
        _handlers[typeof(TEvent)] = (current, _, domainEvent) => current is null ? null : update(current, (TEvent)domainEvent);
}
