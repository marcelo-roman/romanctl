using SampleApp.Domain.SampleItems;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Infrastructure.Persistence.EfCore;

internal sealed class SampleItemRepository(AppDbContext db) : ISampleItemRepository
{
    public void Add(SampleItem sampleItem) => db.Set<SampleItem>().Add(sampleItem);

    public void Remove(SampleItem sampleItem) => db.Set<SampleItem>().Remove(sampleItem);

    public Task<SampleItem?> GetByIdAsync(SampleItemId id, CancellationToken cancellationToken) =>
        db.Set<SampleItem>().FirstOrDefaultAsync(item => item.Id == id, cancellationToken);
}
