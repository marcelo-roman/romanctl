using SampleApp.Catalog.Domain;
using SampleApp.Shared.Domain;

namespace SampleApp.Catalog.Application;

internal sealed record CreateProduct(string Name, decimal Price, string? Currency);

internal sealed class CreateProductHandler(IProductRepository products, TimeProvider clock)
{
    public async Task<Guid> HandleAsync(CreateProduct command, CancellationToken cancellationToken)
    {
        var price = new Money(command.Price, command.Currency ?? Money.DefaultCurrency);
        var product = Product.Create(command.Name, price, clock.GetUtcNow());
        products.Add(product);
        await products.SaveChangesAsync(cancellationToken);
        return product.Id;
    }
}
