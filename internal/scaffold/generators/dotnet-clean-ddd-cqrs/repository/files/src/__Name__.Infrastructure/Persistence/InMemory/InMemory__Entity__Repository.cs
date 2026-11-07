using {{.Name}}.Domain.{{.EntityPlural}};

namespace {{.Name}}.Infrastructure.Persistence.InMemory;

internal sealed class InMemory{{.Entity}}Repository(InMemory{{.Entity}}Store store, InMemoryUnitOfWork unitOfWork) : I{{.Entity}}Repository
{
    public void Add({{.Entity}} {{.EntityCamel}}) =>
        unitOfWork.Track({{.EntityCamel}}, () => store.Items[{{.EntityCamel}}.Id] = {{.EntityCamel}});

    public void Remove({{.Entity}} {{.EntityCamel}}) =>
        unitOfWork.Track({{.EntityCamel}}, () => store.Items.TryRemove({{.EntityCamel}}.Id, out _));

    public Task<{{.Entity}}?> GetByIdAsync({{.Entity}}Id id, CancellationToken cancellationToken)
    {
        if (store.Items.TryGetValue(id, out var item))
            unitOfWork.Track(item, () => store.Items[id] = item);
        return Task.FromResult(item);
    }
}
