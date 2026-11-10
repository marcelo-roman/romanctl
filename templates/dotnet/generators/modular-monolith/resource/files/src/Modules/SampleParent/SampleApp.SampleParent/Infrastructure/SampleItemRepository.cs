using SampleApp.SampleParent.Domain;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.SampleParent.Infrastructure;

internal sealed class SampleItemRepository(SampleParentDbContext db) : ISampleItemRepository
{
    public void Add(SampleItem sampleItem) => db.Set<SampleItem>().Add(sampleItem);

    public void Remove(SampleItem sampleItem) => db.Set<SampleItem>().Remove(sampleItem);

    public Task<SampleItem?> FindAsync(Guid id, CancellationToken cancellationToken) =>
        db.Set<SampleItem>().FirstOrDefaultAsync(sampleItem => sampleItem.Id == id, cancellationToken);

    public async Task<IReadOnlyList<SampleItem>> ListRecentAsync(int limit, CancellationToken cancellationToken) =>
        await db.Set<SampleItem>().AsNoTracking()
            .OrderByDescending(sampleItem => sampleItem.CreatedAt)
            .Take(limit)
            .ToListAsync(cancellationToken);

    public Task SaveChangesAsync(CancellationToken cancellationToken) => db.SaveChangesAsync(cancellationToken);
}
