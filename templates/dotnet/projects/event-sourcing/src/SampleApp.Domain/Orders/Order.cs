using SampleApp.Domain.Common;

namespace SampleApp.Domain.Orders;

public sealed class Order : EventSourcedAggregate
{
    public const string DefaultCurrency = "BRL";

    private readonly List<OrderLine> _lines = [];

    public string CustomerName { get; private set; } = "";
    public OrderStatus Status { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; }

    public IReadOnlyList<OrderLine> Lines => _lines;

    public string Currency => _lines.Count == 0 ? DefaultCurrency : _lines[0].UnitPrice.Currency;

    public Money Total => _lines.Aggregate(new Money(0, Currency), (total, line) => total.Add(line.Subtotal));

    public static Order Create(Guid id, string customerName, DateTimeOffset now)
    {
        if (string.IsNullOrWhiteSpace(customerName))
            throw new DomainException("Customer name is required.");

        var order = new Order { Id = id };
        order.Raise(new OrderCreated(customerName.Trim(), now));
        return order;
    }

    public void AddItem(string productName, int quantity, Money unitPrice)
    {
        if (Status != OrderStatus.Draft)
            throw new DomainException("A placed order cannot be changed.");
        if (string.IsNullOrWhiteSpace(productName))
            throw new DomainException("Product name is required.");
        if (quantity <= 0)
            throw new DomainException("Quantity must be greater than zero.");
        if (_lines.Count > 0 && unitPrice.Currency != Currency)
            throw new DomainException("All order items must use the same currency.");

        Raise(new ItemAdded(productName.Trim(), quantity, unitPrice.Amount, unitPrice.Currency));
    }

    public void Place(DateTimeOffset now)
    {
        if (Status != OrderStatus.Draft)
            throw new DomainException("The order has already been placed.");
        if (_lines.Count == 0)
            throw new DomainException("An order without items cannot be placed.");

        var total = Total;
        Raise(new OrderPlaced(total.Amount, total.Currency, now));
    }

    protected override void When(IDomainEvent domainEvent)
    {
        switch (domainEvent)
        {
            case OrderCreated created:
                Apply(created);
                break;
            case ItemAdded added:
                Apply(added);
                break;
            case OrderPlaced placed:
                Apply(placed);
                break;
            default:
                throw UnknownEvent(domainEvent);
        }
    }

    private void Apply(OrderCreated domainEvent)
    {
        CustomerName = domainEvent.CustomerName;
        Status = OrderStatus.Draft;
        CreatedAt = domainEvent.CreatedAt;
    }

    private void Apply(ItemAdded domainEvent) =>
        _lines.Add(new OrderLine(domainEvent.ProductName, domainEvent.Quantity, new Money(domainEvent.UnitPrice, domainEvent.Currency)));

    private void Apply(OrderPlaced domainEvent) => Status = OrderStatus.Placed;
}
