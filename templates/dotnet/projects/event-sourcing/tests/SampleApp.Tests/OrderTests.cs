using SampleApp.Domain.Common;
using SampleApp.Domain.Orders;

namespace SampleApp.Tests;

public class OrderTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);
    private static readonly OrderCreated Created = new("Ana", Now);
    private static readonly ItemAdded Pens = new("Pen", 2, 1.50m, "BRL");

    private static AggregateScenario<Order> Scenario() => new();

    [Fact]
    public void Create_raises_OrderCreated() =>
        Scenario()
            .WhenCreating(id => Order.Create(id, " Ana ", Now))
            .Then(new OrderCreated("Ana", Now));

    [Fact]
    public void Create_requires_the_customer_name() =>
        Scenario()
            .WhenCreating(id => Order.Create(id, " ", Now))
            .ThenFails("Customer name is required.");

    [Fact]
    public void Add_item_raises_ItemAdded_with_default_currency() =>
        Scenario()
            .Given(Created)
            .When(order => order.AddItem(" Pen ", 2, new Money(1.50m, Order.DefaultCurrency)))
            .Then(Pens);

    [Theory]
    [InlineData("", 1, "Product name is required.")]
    [InlineData("Pen", 0, "Quantity must be greater than zero.")]
    public void Rejects_invalid_item(string product, int quantity, string message) =>
        Scenario()
            .Given(Created)
            .When(order => order.AddItem(product, quantity, new Money(1m, "BRL")))
            .ThenFails(message);

    [Fact]
    public void Does_not_mix_currencies_in_the_same_order() =>
        Scenario()
            .Given(Created, Pens)
            .When(order => order.AddItem("Pen", 1, new Money(1m, "USD")))
            .ThenFails("All order items must use the same currency.");

    [Fact]
    public void Place_sums_the_items() =>
        Scenario()
            .Given(Created, Pens, new ItemAdded("Notebook", 1, 10m, "BRL"))
            .When(order => order.Place(Now))
            .Then(new OrderPlaced(13m, "BRL", Now));

    [Fact]
    public void Does_not_place_order_without_items() =>
        Scenario()
            .Given(Created)
            .When(order => order.Place(Now))
            .ThenFails("An order without items cannot be placed.");

    [Fact]
    public void Placed_order_does_not_accept_items() =>
        Scenario()
            .Given(Created, Pens, new OrderPlaced(3m, "BRL", Now))
            .When(order => order.AddItem("Pencil", 1, new Money(1m, "BRL")))
            .ThenFails("A placed order cannot be changed.");

    [Fact]
    public void Placed_order_is_not_placed_again() =>
        Scenario()
            .Given(Created, Pens, new OrderPlaced(3m, "BRL", Now))
            .When(order => order.Place(Now))
            .ThenFails("The order has already been placed.");

    [Fact]
    public void Rebuilds_state_and_version_from_events()
    {
        var order = new Order();

        order.LoadFromHistory(Guid.NewGuid(), [Created, Pens, new OrderPlaced(3m, "BRL", Now)]);

        Assert.Equal(3, order.Version);
        Assert.Equal(OrderStatus.Placed, order.Status);
        Assert.Equal(new Money(3m, "BRL"), order.Total);
        Assert.Empty(order.PendingEvents);
    }

    [Fact]
    public void Committing_events_advances_the_version()
    {
        var order = Order.Create(Guid.NewGuid(), "Ana", Now);
        order.AddItem("Pen", 1, new Money(1m, "BRL"));

        order.MarkCommitted();

        Assert.Equal(2, order.Version);
        Assert.Empty(order.PendingEvents);
    }

    [Theory]
    [InlineData(-1, "BRL", "Amount cannot be negative.")]
    [InlineData(1, "REAL", "Currency must have 3 letters, like BRL.")]
    public void Rejects_invalid_money(decimal amount, string currency, string message) =>
        Assert.Equal(message, Assert.Throws<DomainException>(() => new Money(amount, currency)).Message);
}
