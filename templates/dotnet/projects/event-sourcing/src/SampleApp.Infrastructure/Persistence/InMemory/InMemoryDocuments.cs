using System.Collections.Concurrent;
using SampleApp.Application.EventSourcing;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal sealed class InMemoryDocuments<TView> where TView : class, IReadModel
{
    private readonly ConcurrentDictionary<Guid, TView> _documents = new();

    public IEnumerable<TView> All => _documents.Values;

    public TView? Find(Guid id) => _documents.GetValueOrDefault(id);

    public void Store(TView document) => _documents[document.Id] = document;

    public void Remove(Guid id) => _documents.TryRemove(id, out _);
}
