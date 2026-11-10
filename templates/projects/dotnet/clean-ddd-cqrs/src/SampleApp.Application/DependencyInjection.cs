using SampleApp.Application.Messaging;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Application;

public static class DependencyInjection
{
    private static readonly Type[] HandlerTypes =
    [
        typeof(ICommandHandler<,>),
        typeof(IQueryHandler<,>),
        typeof(IDomainEventHandler<>),
    ];

    public static IServiceCollection AddApplication(this IServiceCollection services)
    {
        services.AddSingleton(TimeProvider.System);
        services.AddScoped<Dispatcher>();
        services.AddScoped<IDispatcher>(provider => provider.GetRequiredService<Dispatcher>());
        services.AddScoped<IDomainEventPublisher>(provider => provider.GetRequiredService<Dispatcher>());

        var implementations = typeof(DependencyInjection).Assembly.GetTypes()
            .Where(type => type is { IsClass: true, IsAbstract: false });

        foreach (var implementation in implementations)
        foreach (var contract in implementation.GetInterfaces())
        {
            if (contract.IsGenericType && HandlerTypes.Contains(contract.GetGenericTypeDefinition()))
                services.AddScoped(contract, implementation);
        }

        return services;
    }
}
