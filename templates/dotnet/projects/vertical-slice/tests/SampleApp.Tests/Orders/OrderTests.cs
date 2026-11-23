using SampleApp.Api.Domain;
using SampleApp.Api.Domain.Orders;

namespace SampleApp.Tests.Orders;

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
    public void New_order_starts_as_draft_with_zero_total_in_BRL()
    {
        var order = Order.Create(" Ana ", Now);

        Assert.Equal("Ana", order.CustomerName);
        Assert.Equal(OrderStatus.Draft, order.Status);
        Assert.Equal(new Money(0m, Order.DefaultCurrency), order.Total);
    }

    [Fact]
    public void Rejects_customer_without_name() =>
        Assert.Throws<DomainException>(() => Order.Create(" ", Now));

    [Fact]
    public void Does_not_place_order_without_items()
    {
        var order = Order.Create("Ana", Now);

        Assert.Throws<DomainException>(() => order.Place());
    }

    [Fact]
    public void Placing_the_order_prevents_changes()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem("Pen", 2, new Money(1.50m, "BRL"));

        order.Place();

        Assert.Equal(OrderStatus.Placed, order.Status);
        Assert.Throws<DomainException>(() => order.AddItem("Pencil", 1, new Money(1m, "BRL")));
        Assert.Throws<DomainException>(() => order.Place());
    }

    [Fact]
    public void Does_not_mix_currencies_in_the_same_order()
    {
        var order = Order.Create("Ana", Now);
        order.AddItem("Pen", 1, new Money(1m, "BRL"));

        Assert.Throws<DomainException>(() => order.AddItem("Imported pen", 1, new Money(1m, "USD")));
    }

    [Theory]
    [InlineData("", 1)]
    [InlineData("Pen", 0)]
    public void Rejects_invalid_item(string product, int quantity)
    {
        var order = Order.Create("Ana", Now);

        Assert.Throws<DomainException>(() => order.AddItem(product, quantity, new Money(1m, "BRL")));
    }

    [Theory]
    [InlineData(-1, "BRL")]
    [InlineData(1, "REAL")]
    [InlineData(1, "")]
    public void Rejects_invalid_money(decimal amount, string currency) =>
        Assert.Throws<DomainException>(() => new Money(amount, currency));
}
