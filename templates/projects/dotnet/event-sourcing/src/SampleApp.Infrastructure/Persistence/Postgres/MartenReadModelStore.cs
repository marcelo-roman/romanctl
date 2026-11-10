using Marten;
using SampleApp.Application.EventSourcing;

namespace SampleApp.Infrastructure.Persistence.Postgres;

internal sealed class MartenReadModelStore<TView>(IQuerySession session) : IReadModelStore<TView>
    where TView : class, IReadModel
{
    public Task<TView?> FindAsync(Guid id, CancellationToken cancellationToken) =>
        session.LoadAsync<TView>(id, cancellationToken);

    public async Task<IReadOnlyList<TView>> ListAsync(int limit, CancellationToken cancellationToken) =>
        await session.Query<TView>()
            .OrderByDescending(view => view.CreatedAt)
            .Take(limit)
            .ToListAsync(cancellationToken);
}
