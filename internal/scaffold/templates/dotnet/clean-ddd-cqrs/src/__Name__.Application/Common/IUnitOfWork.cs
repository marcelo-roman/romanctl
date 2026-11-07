namespace {{.Name}}.Application.Common;

public interface IUnitOfWork
{
    /// <summary>Grava as alterações e, depois disso, publica os eventos de domínio.</summary>
    Task SaveChangesAsync(CancellationToken cancellationToken);
}
