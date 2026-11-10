using SampleApp.Catalog.Domain;
using SampleApp.Shared.Application;

namespace SampleApp.Catalog.Application;

internal sealed class ListProductsHandler(IProductRepository products)
{
    public async Task<IReadOnlyList<ProductResponse>> HandleAsync(int? limit, CancellationToken cancellationToken)
    {
        var recent = await products.ListRecentAsync(ListLimit.Clamp(limit), cancellationToken);
        return recent.Select(ProductResponse.From).ToList();
    }
}
