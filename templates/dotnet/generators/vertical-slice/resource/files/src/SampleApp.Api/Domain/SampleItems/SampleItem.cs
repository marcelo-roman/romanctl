namespace SampleApp.Api.Domain.SampleItems;

public sealed class SampleItem
{
    public Guid Id { get; private set; }
    public string Name { get; private set; } = default!;
    public DateTimeOffset CreatedAt { get; private set; }

    private SampleItem() { }

    public static SampleItem Create(string name, DateTimeOffset now)
    {
        var sampleItem = new SampleItem { Id = Guid.NewGuid(), CreatedAt = now };
        sampleItem.Rename(name);
        return sampleItem;
    }

    public void Rename(string name)
    {
        if (string.IsNullOrWhiteSpace(name))
            throw new DomainException("O nome é obrigatório.");

        Name = name.Trim();
    }
}
