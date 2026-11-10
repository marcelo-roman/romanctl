using SampleApp.Shared.Domain;

namespace SampleApp.Catalog.Domain;

internal sealed class Product
{
    public Guid Id { get; private set; }
    public string Name { get; private set; } = default!;
    public Money Price { get; private set; } = default!;
    public int SoldQuantity { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; }

    private Product() { }

    public static Product Create(string name, Money price, DateTimeOffset now)
    {
        if (string.IsNullOrWhiteSpace(name))
            throw new DomainException("O nome do produto é obrigatório.");

        return new Product
        {
            Id = Guid.NewGuid(),
            Name = name.Trim(),
            Price = price,
            CreatedAt = now,
        };
    }

    public void RecordSale(int quantity)
    {
        if (quantity <= 0)
            throw new DomainException("A quantidade vendida deve ser maior que zero.");

        SoldQuantity += quantity;
    }
}
