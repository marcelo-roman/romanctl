using SampleApp.Shared.Persistence;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Tests;

public sealed class InMemoryTestDatabase(string name) : IModuleDatabase
{
    public void Configure(DbContextOptionsBuilder options, string schema) =>
        options.UseInMemoryDatabase($"{name}-{schema}");

    public Task InitializeAsync(DbContext context, string schema, CancellationToken cancellationToken) =>
        context.Database.EnsureCreatedAsync(cancellationToken);
}
