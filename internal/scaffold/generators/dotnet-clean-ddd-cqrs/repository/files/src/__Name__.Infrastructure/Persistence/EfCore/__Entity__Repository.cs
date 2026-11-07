using {{.Name}}.Domain.{{.EntityPlural}};
using Microsoft.EntityFrameworkCore;

namespace {{.Name}}.Infrastructure.Persistence.EfCore;

internal sealed class {{.Entity}}Repository(AppDbContext db) : I{{.Entity}}Repository
{
    public void Add({{.Entity}} {{.EntityCamel}}) => db.Set<{{.Entity}}>().Add({{.EntityCamel}});

    public void Remove({{.Entity}} {{.EntityCamel}}) => db.Set<{{.Entity}}>().Remove({{.EntityCamel}});

    public Task<{{.Entity}}?> GetByIdAsync({{.Entity}}Id id, CancellationToken cancellationToken) =>
        db.Set<{{.Entity}}>().FirstOrDefaultAsync(item => item.Id == id, cancellationToken);
}
