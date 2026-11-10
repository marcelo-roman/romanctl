using SampleApp.Catalog.Domain;
using SampleApp.Shared.Domain;

namespace SampleApp.Tests.Catalog;

public class ProductTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Produto_novo_apara_o_nome_e_comeca_sem_vendas()
    {
        var product = Product.Create(" Caneta ", new Money(1.50m, "BRL"), Now);

        Assert.Equal("Caneta", product.Name);
        Assert.Equal(new Money(1.50m, "BRL"), product.Price);
        Assert.Equal(0, product.SoldQuantity);
        Assert.Equal(Now, product.CreatedAt);
    }

    [Fact]
    public void Rejeita_produto_sem_nome() =>
        Assert.Throws<DomainException>(() => Product.Create(" ", new Money(1m, "BRL"), Now));

    [Fact]
    public void Registrar_venda_acumula_a_quantidade()
    {
        var product = Product.Create("Caneta", new Money(1m, "BRL"), Now);

        product.RecordSale(2);
        product.RecordSale(3);

        Assert.Equal(5, product.SoldQuantity);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-1)]
    public void Rejeita_venda_sem_quantidade(int quantity)
    {
        var product = Product.Create("Caneta", new Money(1m, "BRL"), Now);

        Assert.Throws<DomainException>(() => product.RecordSale(quantity));
    }
}
