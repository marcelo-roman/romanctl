namespace SampleApp.Domain.Common;

public abstract class EventSourcedAggregate
{
    private readonly List<IDomainEvent> _pendingEvents = [];

    public Guid Id { get; protected set; }

    public long Version { get; private set; }

    public IReadOnlyList<IDomainEvent> PendingEvents => _pendingEvents;

    public void LoadFromHistory(Guid id, IEnumerable<IDomainEvent> history)
    {
        Id = id;
        foreach (var domainEvent in history)
        {
            When(domainEvent);
            Version++;
        }
    }

    public void MarkCommitted()
    {
        Version += _pendingEvents.Count;
        _pendingEvents.Clear();
    }

    protected void Raise(IDomainEvent domainEvent)
    {
        When(domainEvent);
        _pendingEvents.Add(domainEvent);
    }

    protected abstract void When(IDomainEvent domainEvent);

    protected InvalidOperationException UnknownEvent(IDomainEvent domainEvent) =>
        new($"{GetType().Name} não sabe aplicar o evento {domainEvent.GetType().Name}.");
}
