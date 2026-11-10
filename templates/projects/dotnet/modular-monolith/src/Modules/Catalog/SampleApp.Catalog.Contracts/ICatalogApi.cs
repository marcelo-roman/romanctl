namespace SampleApp.Catalog.Contracts;

public interface ICatalogApi
{
    Task<ProductSnapshot?> FindProductAsync(Guid productId, CancellationToken cancellationToken);
}
