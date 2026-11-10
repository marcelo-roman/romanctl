using SampleApp.Catalog.Domain;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Catalog.Infrastructure;

internal sealed class ProductRepository(CatalogDbContext db) : IProductRepository
{
    public void Add(Product product) => db.Set<Product>().Add(product);

    public Task<Product?> FindAsync(Guid id, CancellationToken cancellationToken) =>
        db.Set<Product>().FirstOrDefaultAsync(product => product.Id == id, cancellationToken);

    public async Task<IReadOnlyList<Product>> FindManyAsync(IReadOnlyCollection<Guid> ids, CancellationToken cancellationToken) =>
        await db.Set<Product>().Where(product => ids.Contains(product.Id)).ToListAsync(cancellationToken);

    public async Task<IReadOnlyList<Product>> ListRecentAsync(int limit, CancellationToken cancellationToken) =>
        await db.Set<Product>().AsNoTracking()
            .OrderByDescending(product => product.CreatedAt)
            .Take(limit)
            .ToListAsync(cancellationToken);

    public Task SaveChangesAsync(CancellationToken cancellationToken) => db.SaveChangesAsync(cancellationToken);
}
