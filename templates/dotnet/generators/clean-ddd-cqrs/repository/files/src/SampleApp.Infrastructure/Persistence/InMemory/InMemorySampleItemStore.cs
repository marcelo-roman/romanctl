using System.Collections.Concurrent;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal sealed class InMemorySampleItemStore
{
    public ConcurrentDictionary<SampleItemId, SampleItem> Items { get; } = new();
}
