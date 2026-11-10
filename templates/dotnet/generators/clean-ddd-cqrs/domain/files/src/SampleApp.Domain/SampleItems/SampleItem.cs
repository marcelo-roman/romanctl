using SampleApp.Domain.Common;

namespace SampleApp.Domain.SampleItems;

public sealed class SampleItem : AggregateRoot<SampleItemId>
{
    public string Name { get; private set; } = default!;
    public DateTimeOffset CreatedAt { get; private set; }

    private SampleItem() { }

    public static SampleItem Create(string name, DateTimeOffset now)
    {
        var sampleItem = new SampleItem { Id = SampleItemId.New(), CreatedAt = now };
        sampleItem.Rename(name);
        sampleItem.Raise(new SampleItemCreated(sampleItem.Id, now));
        return sampleItem;
    }

    public void Rename(string name)
    {
        if (string.IsNullOrWhiteSpace(name))
            throw new DomainException("O nome é obrigatório.");

        Name = name.Trim();
    }
}
