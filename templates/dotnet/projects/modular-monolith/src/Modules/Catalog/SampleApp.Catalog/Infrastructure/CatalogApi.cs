using SampleApp.Catalog.Contracts;
using SampleApp.Catalog.Domain;

namespace SampleApp.Catalog.Infrastructure;

internal sealed class CatalogApi(IProductRepository products) : ICatalogApi
{
    public async Task<ProductSnapshot?> FindProductAsync(Guid productId, CancellationToken cancellationToken)
    {
        var product = await products.FindAsync(productId, cancellationToken);
        if (product is null)
            return null;

        return new ProductSnapshot(product.Id, product.Name, product.Price.Amount, product.Price.Currency);
    }
}
