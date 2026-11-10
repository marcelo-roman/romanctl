using SampleApp.Application.EventSourcing;
using SampleApp.Application.Orders;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Infrastructure.Persistence.InMemory;

public static class InMemoryPersistence
{
    public static IServiceCollection AddInMemoryPersistence(this IServiceCollection services)
    {
        services.AddSingleton(typeof(InMemoryDocuments<>));
        services.AddSingleton(typeof(IReadModelStore<>), typeof(InMemoryReadModelStore<>));
        services.AddSingleton<IEventStore, InMemoryEventStore>();

        services.AddInMemoryProjection(new OrderSummaryProjection());
        // romanctl:projections
        return services;
    }

    private static void AddInMemoryProjection<TView>(this IServiceCollection services, Projection<TView> projection)
        where TView : class, IReadModel =>
        services.AddSingleton<IInMemoryProjection>(provider =>
            new InMemoryProjection<TView>(projection, provider.GetRequiredService<InMemoryDocuments<TView>>()));
}
