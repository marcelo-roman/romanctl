namespace SampleApp.Application.Common;

public interface IUnitOfWork
{
    /// <summary>Saves the changes and only then publishes the domain events.</summary>
    Task SaveChangesAsync(CancellationToken cancellationToken);
}
