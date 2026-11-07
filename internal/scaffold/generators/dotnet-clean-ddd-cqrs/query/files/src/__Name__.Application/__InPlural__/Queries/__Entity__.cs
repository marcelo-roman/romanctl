using {{.Name}}.Application.Messaging;

namespace {{.Name}}.Application.{{.InPlural}}.Queries;

public sealed record {{.Entity}}(int Limit = 50) : IQuery<IReadOnlyList<{{.In}}Dto>>;

internal sealed class {{.Entity}}Handler(I{{.In}}ReadStore store) : IQueryHandler<{{.Entity}}, IReadOnlyList<{{.In}}Dto>>
{
    public Task<IReadOnlyList<{{.In}}Dto>> HandleAsync({{.Entity}} query, CancellationToken cancellationToken) =>
        store.ListAsync(Math.Clamp(query.Limit, 1, 200), cancellationToken);
}
