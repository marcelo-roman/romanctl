using {{.Name}}.Application.Common;
using {{.Name}}.Application.Orders;
using {{.Name}}.Domain.Orders;
{{- if .Postgres}}
using {{.Name}}.Infrastructure.Persistence.EfCore;
using Microsoft.EntityFrameworkCore;
{{- else}}
using {{.Name}}.Infrastructure.Persistence.InMemory;
{{- end}}
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace {{.Name}}.Infrastructure;

public static class DependencyInjection
{
    public static IServiceCollection AddInfrastructure(this IServiceCollection services, IConfiguration configuration)
    {
{{- if .Postgres}}
        var connectionString = configuration.GetConnectionString("Default")
            ?? throw new InvalidOperationException("Falta a connection string Default.");

        services.AddDbContext<AppDbContext>(options => options.UseNpgsql(connectionString));
        services.AddScoped<IUnitOfWork>(provider => provider.GetRequiredService<AppDbContext>());
        services.AddScoped<IOrderRepository, OrderRepository>();
        services.AddScoped<IOrderReadStore, OrderReadStore>();
{{- else}}
        services.AddSingleton<InMemoryDatabase>();
        services.AddScoped<InMemoryOrderRepository>();
        services.AddScoped<IOrderRepository>(provider => provider.GetRequiredService<InMemoryOrderRepository>());
        services.AddScoped<IUnitOfWork>(provider => provider.GetRequiredService<InMemoryOrderRepository>());
        services.AddScoped<IOrderReadStore, InMemoryOrderReadStore>();
{{- end}}
        return services;
    }
}
