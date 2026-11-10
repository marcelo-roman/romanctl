using Microsoft.EntityFrameworkCore;

namespace SampleApp.Shared.Persistence;

public interface IModuleDatabase
{
    void Configure(DbContextOptionsBuilder options, string schema);

    Task InitializeAsync(DbContext context, string schema, CancellationToken cancellationToken);
}
