using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;

namespace {{.Name}}.Application.{{.EntityPlural}}.Queries;

public sealed record Get{{.Entity}}ById(Guid Id) : IQuery<{{.Entity}}Dto>;

internal sealed class Get{{.Entity}}ByIdHandler(I{{.Entity}}ReadStore store) : IQueryHandler<Get{{.Entity}}ById, {{.Entity}}Dto>
{
    public async Task<{{.Entity}}Dto> HandleAsync(Get{{.Entity}}ById query, CancellationToken cancellationToken) =>
        await store.FindAsync(query.Id, cancellationToken)
        ?? throw new NotFoundException($"{{.Entity}} {query.Id} não encontrado.");
}
