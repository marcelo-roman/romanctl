using SampleApp.Domain.Common;
using SampleApp.Domain.Orders;

namespace SampleApp.Tests;

public class OrderTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Total_sums_item_subtotals()
    {
        var order = Order.Create("Ana", Now);

        order.AddItem("Pen", 2, new Money(1.50m, "BRL"));
        order.AddItem("Notebook", 1, new Money(10m, "BRL"));

        Assert.Equal(new Money(13m, "BRL"), order.Total);
    }

    [Fact]
    public void Does_not_place_order_without_items()
    {
        var order = Order.Create("Ana", Now);

        Assert.Throws<DomainException>(() => order.Place(Now));
    }

    [Fact]
    public void Placing_the_order_raises_event_and_prevents_changes()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem("Pen", 2, new Money(1.50m, "BRL"));

        order.Place(Now);

        var placed = Assert.IsType<OrderPlaced>(Assert.Single(order.DomainEvents));
        Assert.Equal(3m, placed.Total);
        Assert.Throws<DomainException>(() => order.AddItem("Pencil", 1, new Money(1m, "BRL")));
        Assert.Throws<DomainException>(() => order.Place(Now));
    }

    [Fact]
    public void Does_not_mix_currencies_in_the_same_order()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem("Pen", 1, new Money(1m, "BRL"));

        Assert.Throws<DomainException>(() => order.AddItem("Eraser", 1, new Money(1m, "USD")));
    }

    [Theory]
    [InlineData("", 1)]
    [InlineData("Pen", 0)]
    public void Rejects_invalid_item(string product, int quantity)
    {
        var order = Order.Create("Ana", Now);

        Assert.Throws<DomainException>(() => order.AddItem(product, quantity, new Money(1m, "BRL")));
    }

    [Fact]
    public void Rejects_negative_amount() =>
        Assert.Throws<DomainException>(() => new Money(-1m, "BRL"));
}
