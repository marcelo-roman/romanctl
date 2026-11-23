using SampleApp.Ordering.Domain;
using SampleApp.Shared.Domain;

namespace SampleApp.Tests.Ordering;

public class OrderTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);
    private static readonly Guid Pen = Guid.NewGuid();
    private static readonly Guid Notebook = Guid.NewGuid();

    [Fact]
    public void Total_sums_item_subtotals()
    {
        var order = Order.Create("Ana", Now);

        order.AddItem(Pen, "Pen", 2, new Money(1.50m, "BRL"));
        order.AddItem(Notebook, "Notebook", 1, new Money(10m, "BRL"));

        Assert.Equal(new Money(13m, "BRL"), order.Total);
    }

    [Fact]
    public void New_order_starts_as_draft_with_zero_total_in_BRL()
    {
        var order = Order.Create(" Ana ", Now);

        Assert.Equal("Ana", order.CustomerName);
        Assert.Equal(OrderStatus.Draft, order.Status);
        Assert.Equal(Money.Zero(Money.DefaultCurrency), order.Total);
        Assert.Null(order.PlacedAt);
    }

    [Fact]
    public void Rejects_customer_without_name() =>
        Assert.Throws<DomainException>(() => Order.Create(" ", Now));

    [Fact]
    public void Does_not_place_order_without_items()
    {
        var order = Order.Create("Ana", Now);

        Assert.Throws<DomainException>(() => order.Place(Now));
    }

    [Fact]
    public void Placing_the_order_prevents_changes()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem(Pen, "Pen", 2, new Money(1.50m, "BRL"));

        order.Place(Now.AddMinutes(1));

        Assert.Equal(OrderStatus.Placed, order.Status);
        Assert.Equal(Now.AddMinutes(1), order.PlacedAt);
        Assert.Throws<DomainException>(() => order.AddItem(Notebook, "Notebook", 1, new Money(1m, "BRL")));
        Assert.Throws<DomainException>(() => order.Place(Now));
    }

    [Fact]
    public void Does_not_mix_currencies_in_the_same_order()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem(Pen, "Pen", 1, new Money(1m, "BRL"));

        Assert.Throws<DomainException>(() => order.AddItem(Notebook, "Imported notebook", 1, new Money(1m, "USD")));
    }

    [Theory]
    [InlineData(true, "", 1)]
    [InlineData(true, "Pen", 0)]
    [InlineData(false, "Pen", 1)]
    public void Rejects_invalid_item(bool withProduct, string productName, int quantity)
    {
        var order = Order.Create("Ana", Now);
        var productId = withProduct ? Pen : Guid.Empty;

        Assert.Throws<DomainException>(() => order.AddItem(productId, productName, quantity, new Money(1m, "BRL")));
    }
}
