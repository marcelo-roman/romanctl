using SampleApp.Domain.Common;

namespace SampleApp.Domain.SampleItems;

public sealed class SampleItem : EventSourcedAggregate
{
    public string Name { get; private set; } = "";
    public DateTimeOffset CreatedAt { get; private set; }

    public static SampleItem Create(Guid id, string name, DateTimeOffset now)
    {
        var sampleItem = new SampleItem { Id = id };
        sampleItem.Raise(new SampleItemCreated(RequireName(name), now));
        return sampleItem;
    }

    public void Rename(string name)
    {
        var newName = RequireName(name);
        if (newName == Name)
            return;

        Raise(new SampleItemRenamed(newName));
    }

    private static string RequireName(string name) =>
        string.IsNullOrWhiteSpace(name) ? throw new DomainException("O nome é obrigatório.") : name.Trim();

    protected override void When(IDomainEvent domainEvent)
    {
        switch (domainEvent)
        {
            case SampleItemCreated created:
                Apply(created);
                break;
            case SampleItemRenamed renamed:
                Apply(renamed);
                break;
            default:
                throw UnknownEvent(domainEvent);
        }
    }

    private void Apply(SampleItemCreated domainEvent)
    {
        Name = domainEvent.Name;
        CreatedAt = domainEvent.CreatedAt;
    }

    private void Apply(SampleItemRenamed domainEvent) => Name = domainEvent.Name;
}
