using SampleApp.Application.Messaging;
using SampleApp.Domain.Orders;
using Microsoft.Extensions.Logging;

namespace SampleApp.Application.Orders.EventHandlers;

internal sealed class LogOrderPlaced(ILogger<LogOrderPlaced> logger) : IDomainEventHandler<OrderPlaced>
{
    public Task HandleAsync(OrderPlaced domainEvent, CancellationToken cancellationToken)
    {
        logger.LogInformation("Pedido {OrderId} realizado: {Total} {Currency}",
            domainEvent.OrderId.Value, domainEvent.Total, domainEvent.Currency);
        return Task.CompletedTask;
    }
}
