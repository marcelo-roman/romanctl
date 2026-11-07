using {{.Name}}.Application.Messaging;

namespace {{.Name}}.Application.{{.EntityPlural}}.Queries;

public sealed record List{{.EntityPlural}}(int Limit = 50) : IQuery<IReadOnlyList<{{.Entity}}Dto>>;

internal sealed class List{{.EntityPlural}}Handler(I{{.Entity}}ReadStore store) : IQueryHandler<List{{.EntityPlural}}, IReadOnlyList<{{.Entity}}Dto>>
{
    public Task<IReadOnlyList<{{.Entity}}Dto>> HandleAsync(List{{.EntityPlural}} query, CancellationToken cancellationToken) =>
        store.ListAsync(Math.Clamp(query.Limit, 1, 200), cancellationToken);
}
