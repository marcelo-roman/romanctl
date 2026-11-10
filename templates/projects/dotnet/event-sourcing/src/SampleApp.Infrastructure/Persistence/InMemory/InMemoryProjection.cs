using SampleApp.Application.EventSourcing;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal sealed class InMemoryProjection<TView>(Projection<TView> projection, InMemoryDocuments<TView> documents)
    : IInMemoryProjection where TView : class, IReadModel
{
    public void Apply(StoredEvent stored)
    {
        if (!projection.EventTypes.Contains(stored.Data.GetType()))
            return;

        var current = documents.Find(stored.StreamId);
        var next = projection.Evolve(current, stored.StreamId, stored.Data);
        if (next is null)
        {
            documents.Remove(stored.StreamId);
            return;
        }

        documents.Store(next);
    }
}
