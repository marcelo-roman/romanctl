using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

namespace SampleApp.Shared.Messaging;

internal sealed class IntegrationEventDispatcher(
    InProcessEventBus bus,
    IServiceScopeFactory scopes,
    ILogger<IntegrationEventDispatcher> logger) : BackgroundService
{
    private const string HandleMethod = nameof(IIntegrationEventHandler<IIntegrationEvent>.HandleAsync);

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        await foreach (var integrationEvent in bus.ReadAllAsync(stoppingToken))
            await DispatchAsync(integrationEvent, stoppingToken);
    }

    private async Task DispatchAsync(IIntegrationEvent integrationEvent, CancellationToken cancellationToken)
    {
        var handlerType = typeof(IIntegrationEventHandler<>).MakeGenericType(integrationEvent.GetType());
        var handle = handlerType.GetMethod(HandleMethod)!;

        await using var scope = scopes.CreateAsyncScope();
        foreach (var handler in scope.ServiceProvider.GetServices(handlerType))
        {
            try
            {
                await (Task)handle.Invoke(handler, [integrationEvent, cancellationToken])!;
            }
            catch (Exception exception) when (exception is not OperationCanceledException)
            {
                logger.LogError(exception, "Falha ao tratar o evento {Event} em {Handler}.",
                    integrationEvent.GetType().Name, handler!.GetType().Name);
            }
        }
    }
}
