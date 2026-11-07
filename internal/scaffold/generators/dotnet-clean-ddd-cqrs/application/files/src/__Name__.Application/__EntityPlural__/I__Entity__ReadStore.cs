namespace {{.Name}}.Application.{{.EntityPlural}};

/// <summary>Lado de leitura do CQRS: devolve DTOs e não passa pelo agregado para alterar nada.</summary>
public interface I{{.Entity}}ReadStore
{
    Task<{{.Entity}}Dto?> FindAsync(Guid id, CancellationToken cancellationToken);
    Task<IReadOnlyList<{{.Entity}}Dto>> ListAsync(int limit, CancellationToken cancellationToken);
}
