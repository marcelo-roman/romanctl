using SampleApp.Shared.Domain;

namespace SampleApp.Ordering.Domain;

internal sealed class Order
{
    private readonly List<OrderItem> _items = [];

    public Guid Id { get; private set; }
    public string CustomerName { get; private set; } = default!;
    public OrderStatus Status { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; }
    public DateTimeOffset? PlacedAt { get; private set; }

    public IReadOnlyCollection<OrderItem> Items => _items;

    public Money Total => _items.Aggregate(
        Money.Zero(_items.FirstOrDefault()?.UnitPrice.Currency ?? Money.DefaultCurrency),
        (total, item) => total.Add(item.Subtotal));

    private Order() { }

    public static Order Create(string customerName, DateTimeOffset now)
    {
        if (string.IsNullOrWhiteSpace(customerName))
            throw new DomainException("Customer name is required.");

        return new Order
        {
            Id = Guid.NewGuid(),
            CustomerName = customerName.Trim(),
            Status = OrderStatus.Draft,
            CreatedAt = now,
        };
    }

    public void AddItem(Guid productId, string productName, int quantity, Money unitPrice)
    {
        if (Status != OrderStatus.Draft)
            throw new DomainException("A placed order cannot be changed.");
        if (_items.Count > 0 && _items[0].UnitPrice.Currency != unitPrice.Currency)
            throw new DomainException("All order items must use the same currency.");

        _items.Add(new OrderItem(productId, productName, quantity, unitPrice));
    }

    public void Place(DateTimeOffset now)
    {
        if (Status != OrderStatus.Draft)
            throw new DomainException("The order has already been placed.");
        if (_items.Count == 0)
            throw new DomainException("An order without items cannot be placed.");

        Status = OrderStatus.Placed;
        PlacedAt = now;
    }
}
