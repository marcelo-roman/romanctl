using SampleApp.Shared.Persistence.Postgres;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

namespace SampleApp.Shared.Persistence;

public static class ModuleDatabaseRegistration
{
    public static IHostApplicationBuilder AddModuleDatabase(this IHostApplicationBuilder builder)
    {
        var connectionString = builder.Configuration.GetConnectionString("Default")
            ?? throw new InvalidOperationException("The Default connection string is missing.");

        builder.Services.AddSingleton<IModuleDatabase>(new PostgresModuleDatabase(connectionString));
        return builder;
    }
}
