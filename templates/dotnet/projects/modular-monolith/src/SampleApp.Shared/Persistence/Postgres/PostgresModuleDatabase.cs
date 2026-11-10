using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Infrastructure;
using Microsoft.EntityFrameworkCore.Storage;

namespace SampleApp.Shared.Persistence.Postgres;

internal sealed class PostgresModuleDatabase(string connectionString) : IModuleDatabase
{
    public void Configure(DbContextOptionsBuilder options, string schema) =>
        options.UseNpgsql(connectionString, npgsql => npgsql.MigrationsHistoryTable("__EFMigrationsHistory", schema));

    public async Task InitializeAsync(DbContext context, string schema, CancellationToken cancellationToken)
    {
        var creator = context.GetService<IRelationalDatabaseCreator>();
        if (!await creator.ExistsAsync(cancellationToken))
            await creator.CreateAsync(cancellationToken);
        if (await SchemaHasTablesAsync(context, schema, cancellationToken))
            return;

        await creator.CreateTablesAsync(cancellationToken);
    }

    private static Task<bool> SchemaHasTablesAsync(DbContext context, string schema, CancellationToken cancellationToken) =>
        context.Database
            .SqlQuery<bool>($"select exists (select 1 from information_schema.tables where table_schema = {schema}) as \"Value\"")
            .SingleAsync(cancellationToken);
}
