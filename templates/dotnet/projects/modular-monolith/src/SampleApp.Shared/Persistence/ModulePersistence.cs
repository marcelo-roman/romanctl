using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Shared.Persistence;

public static class ModulePersistence
{
    public static IServiceCollection AddModuleDbContext<TContext>(this IServiceCollection services, string schema)
        where TContext : DbContext
    {
        services.AddDbContext<TContext>((provider, options) =>
            provider.GetRequiredService<IModuleDatabase>().Configure(options, schema));
        services.AddSingleton(new ModuleSchema(typeof(TContext), schema));
        return services;
    }

    public static async Task InitializeModuleDatabasesAsync(this IServiceProvider services, CancellationToken cancellationToken = default)
    {
        await using var scope = services.CreateAsyncScope();
        var database = scope.ServiceProvider.GetRequiredService<IModuleDatabase>();
        foreach (var module in scope.ServiceProvider.GetServices<ModuleSchema>())
        {
            var context = (DbContext)scope.ServiceProvider.GetRequiredService(module.ContextType);
            await database.InitializeAsync(context, module.Schema, cancellationToken);
        }
    }
}
