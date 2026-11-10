using Marten;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Orders;
using SampleApp.Infrastructure.Persistence.Postgres;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Infrastructure;

public static class DependencyInjection
{
    public static IServiceCollection AddInfrastructure(this IServiceCollection services, IConfiguration configuration)
    {
        var connectionString = configuration.GetConnectionString("Default")
            ?? throw new InvalidOperationException("Falta a connection string Default.");

        services.AddMarten(options =>
        {
            options.UseDefaults(connectionString);
            options.AddInlineProjection(new OrderSummaryProjection());
            // romanctl:projections
        }).UseLightweightSessions();

        services.AddScoped<IEventStore, MartenEventStore>();
        services.AddScoped(typeof(IReadModelStore<>), typeof(MartenReadModelStore<>));
        return services;
    }
}
