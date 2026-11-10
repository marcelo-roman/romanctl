using SampleApp.Application.SampleItems;
using SampleApp.Domain.SampleItems;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Infrastructure.Persistence.EfCore;

internal sealed class SampleItemReadStore(AppDbContext db) : ISampleItemReadStore
{
    public async Task<SampleItemDto?> FindAsync(Guid id, CancellationToken cancellationToken)
    {
        var typedId = new SampleItemId(id);
        var item = await db.Set<SampleItem>().AsNoTracking().FirstOrDefaultAsync(x => x.Id == typedId, cancellationToken);
        return item is null ? null : SampleItemDto.From(item);
    }

    public async Task<IReadOnlyList<SampleItemDto>> ListAsync(int limit, CancellationToken cancellationToken)
    {
        var items = await db.Set<SampleItem>().AsNoTracking()
            .OrderByDescending(x => x.CreatedAt)
            .Take(limit)
            .ToListAsync(cancellationToken);
        return items.Select(SampleItemDto.From).ToList();
    }
}
