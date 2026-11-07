using System.Collections.Concurrent;
using {{.Name}}.Domain.{{.EntityPlural}};

namespace {{.Name}}.Infrastructure.Persistence.InMemory;

internal sealed class InMemory{{.Entity}}Store
{
    public ConcurrentDictionary<{{.Entity}}Id, {{.Entity}}> Items { get; } = new();
}
