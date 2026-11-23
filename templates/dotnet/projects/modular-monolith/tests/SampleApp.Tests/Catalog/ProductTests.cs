using SampleApp.Catalog.Domain;
using SampleApp.Shared.Domain;

namespace SampleApp.Tests.Catalog;

public class ProductTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void New_product_trims_the_name_and_starts_without_sales()
    {
        var product = Product.Create(" Pen ", new Money(1.50m, "BRL"), Now);

        Assert.Equal("Pen", product.Name);
        Assert.Equal(new Money(1.50m, "BRL"), product.Price);
        Assert.Equal(0, product.SoldQuantity);
        Assert.Equal(Now, product.CreatedAt);
    }

    [Fact]
    public void Rejects_product_without_name() =>
        Assert.Throws<DomainException>(() => Product.Create(" ", new Money(1m, "BRL"), Now));

    [Fact]
    public void Recording_a_sale_accumulates_the_quantity()
    {
        var product = Product.Create("Pen", new Money(1m, "BRL"), Now);

        product.RecordSale(2);
        product.RecordSale(3);

        Assert.Equal(5, product.SoldQuantity);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-1)]
    public void Rejects_sale_without_quantity(int quantity)
    {
        var product = Product.Create("Pen", new Money(1m, "BRL"), Now);

        Assert.Throws<DomainException>(() => product.RecordSale(quantity));
    }
}
