using SampleApp.Catalog.Domain;
using SampleApp.Shared.Application;

namespace SampleApp.Catalog.Application;

internal sealed class GetProductHandler(IProductRepository products)
{
    public async Task<ProductResponse> HandleAsync(Guid id, CancellationToken cancellationToken)
    {
        var product = await products.FindAsync(id, cancellationToken)
            ?? throw new NotFoundException($"Product {id} not found.");
        return ProductResponse.From(product);
    }
}
