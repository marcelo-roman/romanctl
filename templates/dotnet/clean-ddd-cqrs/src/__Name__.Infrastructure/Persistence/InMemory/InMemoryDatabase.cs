using System.Collections.Concurrent;
using {{.Name}}.Domain.Orders;

namespace {{.Name}}.Infrastructure.Persistence.InMemory;

internal sealed class InMemoryDatabase
{
    public ConcurrentDictionary<OrderId, Order> Orders { get; } = new();
}
