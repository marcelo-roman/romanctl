using System.Collections.Concurrent;
using SampleApp.Domain.Orders;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal sealed class InMemoryDatabase
{
    public ConcurrentDictionary<OrderId, Order> Orders { get; } = new();
}
