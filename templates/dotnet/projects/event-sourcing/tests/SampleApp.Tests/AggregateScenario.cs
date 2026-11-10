using SampleApp.Domain.Common;

namespace SampleApp.Tests;

internal sealed class AggregateScenario<TAggregate> where TAggregate : EventSourcedAggregate, new()
{
    private IDomainEvent[] _history = [];
    private Func<TAggregate> _act = () => throw new InvalidOperationException("Faltou o When.");

    public Guid Id { get; } = Guid.NewGuid();

    public AggregateScenario<TAggregate> Given(params IDomainEvent[] history)
    {
        _history = history;
        return this;
    }

    public AggregateScenario<TAggregate> WhenCreating(Func<Guid, TAggregate> create)
    {
        _act = () => create(Id);
        return this;
    }

    public AggregateScenario<TAggregate> When(Action<TAggregate> command)
    {
        _act = () =>
        {
            var aggregate = new TAggregate();
            aggregate.LoadFromHistory(Id, _history);
            command(aggregate);
            return aggregate;
        };
        return this;
    }

    public void Then(params IDomainEvent[] expected) =>
        Assert.Equal(expected, _act().PendingEvents);

    public void ThenFails(string message) =>
        Assert.Equal(message, Assert.Throws<DomainException>(() => _act()).Message);
}
