using {{.Name}}.Application.{{.EntityPlural}};
using {{.Name}}.Domain.{{.EntityPlural}};
using Microsoft.EntityFrameworkCore;

namespace {{.Name}}.Infrastructure.Persistence.EfCore;

internal sealed class {{.Entity}}ReadStore(AppDbContext db) : I{{.Entity}}ReadStore
{
    public async Task<{{.Entity}}Dto?> FindAsync(Guid id, CancellationToken cancellationToken)
    {
        var typedId = new {{.Entity}}Id(id);
        var item = await db.Set<{{.Entity}}>().AsNoTracking().FirstOrDefaultAsync(x => x.Id == typedId, cancellationToken);
        return item is null ? null : {{.Entity}}Dto.From(item);
    }

    public async Task<IReadOnlyList<{{.Entity}}Dto>> ListAsync(int limit, CancellationToken cancellationToken)
    {
        var items = await db.Set<{{.Entity}}>().AsNoTracking()
            .OrderByDescending(x => x.CreatedAt)
            .Take(limit)
            .ToListAsync(cancellationToken);
        return items.Select({{.Entity}}Dto.From).ToList();
    }
}
