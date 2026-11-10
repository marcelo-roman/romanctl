using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Shared.Messaging;

public static class MessagingRegistration
{
    public static IServiceCollection AddEventBus(this IServiceCollection services)
    {
        services.AddSingleton<InProcessEventBus>();
        services.AddSingleton<IEventBus>(provider => provider.GetRequiredService<InProcessEventBus>());
        services.AddHostedService<IntegrationEventDispatcher>();
        return services;
    }

    public static IServiceCollection AddIntegrationEventHandler<TEvent, THandler>(this IServiceCollection services)
        where TEvent : IIntegrationEvent
        where THandler : class, IIntegrationEventHandler<TEvent> =>
        services.AddScoped<IIntegrationEventHandler<TEvent>, THandler>();
}
