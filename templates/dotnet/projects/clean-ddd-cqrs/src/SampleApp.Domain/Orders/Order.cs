using SampleApp.Domain.Common;

namespace SampleApp.Domain.Orders;

public sealed class Order : AggregateRoot<OrderId>
{
    public const string DefaultCurrency = "BRL";

    private readonly List<OrderItem> _items = [];

    public string CustomerName { get; private set; } = default!;
    public OrderStatus Status { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; }

    public IReadOnlyCollection<OrderItem> Items => _items;

    public Money Total => _items.Aggregate(
        new Money(0, _items.FirstOrDefault()?.UnitPrice.Currency ?? DefaultCurrency),
        (total, item) => total.Add(item.Subtotal));

    private Order() { }

    public static Order Create(string customerName, DateTimeOffset now)
    {
        if (string.IsNullOrWhiteSpace(customerName))
            throw new DomainException("O nome do cliente é obrigatório.");

        return new Order
        {
            Id = OrderId.New(),
            CustomerName = customerName.Trim(),
            Status = OrderStatus.Draft,
            CreatedAt = now,
        };
    }

    public void AddItem(string productName, int quantity, Money unitPrice)
    {
        if (Status != OrderStatus.Draft)
            throw new DomainException("Um pedido já realizado não pode ser alterado.");
        if (_items.Count > 0 && _items[0].UnitPrice.Currency != unitPrice.Currency)
            throw new DomainException("Todos os itens do pedido devem usar a mesma moeda.");

        _items.Add(new OrderItem(productName, quantity, unitPrice));
    }

    public void Place(DateTimeOffset now)
    {
        if (Status != OrderStatus.Draft)
            throw new DomainException("O pedido já foi realizado.");
        if (_items.Count == 0)
            throw new DomainException("Um pedido sem itens não pode ser realizado.");

        Status = OrderStatus.Placed;
        var total = Total;
        Raise(new OrderPlaced(Id, total.Amount, total.Currency, now));
    }
}
