namespace SampleApp.Application.EventSourcing;

public interface IReadModelStore<TView> where TView : class, IReadModel
{
    Task<TView?> FindAsync(Guid id, CancellationToken cancellationToken);
    Task<IReadOnlyList<TView>> ListAsync(int limit, CancellationToken cancellationToken);
}
