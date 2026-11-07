namespace {{.Name}}.Domain.{{.EntityPlural}};

public interface I{{.Entity}}Repository
{
    void Add({{.Entity}} {{.EntityCamel}});
    void Remove({{.Entity}} {{.EntityCamel}});
    Task<{{.Entity}}?> GetByIdAsync({{.Entity}}Id id, CancellationToken cancellationToken);
}
