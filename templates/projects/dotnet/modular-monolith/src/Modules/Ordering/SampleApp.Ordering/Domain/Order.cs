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
            throw new DomainException("O nome do cliente é obrigatório.");

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
            throw new DomainException("Um pedido já realizado não pode ser alterado.");
        if (_items.Count > 0 && _items[0].UnitPrice.Currency != unitPrice.Currency)
            throw new DomainException("Todos os itens do pedido devem usar a mesma moeda.");

        _items.Add(new OrderItem(productId, productName, quantity, unitPrice));
    }

    public void Place(DateTimeOffset now)
    {
        if (Status != OrderStatus.Draft)
            throw new DomainException("O pedido já foi realizado.");
        if (_items.Count == 0)
            throw new DomainException("Um pedido sem itens não pode ser realizado.");

        Status = OrderStatus.Placed;
        PlacedAt = now;
    }
}
