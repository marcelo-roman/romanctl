using Microsoft.EntityFrameworkCore;

namespace SampleApp.Shared.Persistence.InMemory;

internal sealed class InMemoryModuleDatabase(string name) : IModuleDatabase
{
    public void Configure(DbContextOptionsBuilder options, string schema) =>
        options.UseInMemoryDatabase($"{name}-{schema}");

    public Task InitializeAsync(DbContext context, string schema, CancellationToken cancellationToken) =>
        context.Database.EnsureCreatedAsync(cancellationToken);
}
