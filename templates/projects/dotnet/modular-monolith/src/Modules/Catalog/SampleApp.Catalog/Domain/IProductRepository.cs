namespace SampleApp.Catalog.Domain;

internal interface IProductRepository
{
    void Add(Product product);

    Task<Product?> FindAsync(Guid id, CancellationToken cancellationToken);

    Task<IReadOnlyList<Product>> FindManyAsync(IReadOnlyCollection<Guid> ids, CancellationToken cancellationToken);

    Task<IReadOnlyList<Product>> ListRecentAsync(int limit, CancellationToken cancellationToken);

    Task SaveChangesAsync(CancellationToken cancellationToken);
}
