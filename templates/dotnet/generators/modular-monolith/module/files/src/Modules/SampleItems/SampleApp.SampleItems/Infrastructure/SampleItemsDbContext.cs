using Microsoft.EntityFrameworkCore;

namespace SampleApp.SampleItems.Infrastructure;

internal sealed class SampleItemsDbContext(DbContextOptions<SampleItemsDbContext> options) : DbContext(options)
{
    public const string Schema = "sample_items";

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.HasDefaultSchema(Schema);
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(SampleItemsDbContext).Assembly);
    }
}
