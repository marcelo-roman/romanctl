using SampleApp.Application.Common;
using SampleApp.Application.Orders;
using SampleApp.Domain.Orders;
using SampleApp.Infrastructure.Persistence.InMemory;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Infrastructure;

public static class DependencyInjection
{
    public static IServiceCollection AddInfrastructure(this IServiceCollection services, IConfiguration configuration)
    {
        services.AddSingleton<InMemoryDatabase>();
        services.AddScoped<InMemoryUnitOfWork>();
        services.AddScoped<IUnitOfWork>(provider => provider.GetRequiredService<InMemoryUnitOfWork>());
        services.AddScoped<IOrderRepository, InMemoryOrderRepository>();
        services.AddScoped<IOrderReadStore, InMemoryOrderReadStore>();

        // romanctl:repositories
        return services;
    }
}
