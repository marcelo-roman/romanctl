using SampleApp.Application.EventSourcing;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal sealed class InMemoryReadModelStore<TView>(InMemoryDocuments<TView> documents) : IReadModelStore<TView>
    where TView : class, IReadModel
{
    public Task<TView?> FindAsync(Guid id, CancellationToken cancellationToken) =>
        Task.FromResult(documents.Find(id));

    public Task<IReadOnlyList<TView>> ListAsync(int limit, CancellationToken cancellationToken)
    {
        IReadOnlyList<TView> views = documents.All
            .OrderByDescending(view => view.CreatedAt)
            .Take(limit)
            .ToList();
        return Task.FromResult(views);
    }
}
