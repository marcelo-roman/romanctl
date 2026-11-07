using {{.Name}}.Application.{{.EntityPlural}};
using {{.Name}}.Domain.{{.EntityPlural}};

namespace {{.Name}}.Infrastructure.Persistence.InMemory;

internal sealed class InMemory{{.Entity}}ReadStore(InMemory{{.Entity}}Store store) : I{{.Entity}}ReadStore
{
    public Task<{{.Entity}}Dto?> FindAsync(Guid id, CancellationToken cancellationToken) =>
        Task.FromResult(store.Items.TryGetValue(new {{.Entity}}Id(id), out var item) ? {{.Entity}}Dto.From(item) : null);

    public Task<IReadOnlyList<{{.Entity}}Dto>> ListAsync(int limit, CancellationToken cancellationToken)
    {
        IReadOnlyList<{{.Entity}}Dto> items = store.Items.Values
            .OrderByDescending(x => x.CreatedAt)
            .Take(limit)
            .Select({{.Entity}}Dto.From)
            .ToList();
        return Task.FromResult(items);
    }
}
