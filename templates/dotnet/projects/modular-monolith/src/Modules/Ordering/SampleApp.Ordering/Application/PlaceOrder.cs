using SampleApp.Ordering.Contracts;
using SampleApp.Ordering.Domain;
using SampleApp.Shared.Application;
using SampleApp.Shared.Messaging;

namespace SampleApp.Ordering.Application;

internal sealed record PlaceOrder(Guid OrderId);

internal sealed class PlaceOrderHandler(IOrderRepository orders, IEventBus bus, TimeProvider clock)
{
    public async Task HandleAsync(PlaceOrder command, CancellationToken cancellationToken)
    {
        var order = await orders.FindAsync(command.OrderId, cancellationToken)
            ?? throw new NotFoundException($"Pedido {command.OrderId} não encontrado.");

        var now = clock.GetUtcNow();
        order.Place(now);
        await orders.SaveChangesAsync(cancellationToken);
        await bus.PublishAsync(ToIntegrationEvent(order, now), cancellationToken);
    }

    private static OrderPlaced ToIntegrationEvent(Order order, DateTimeOffset now)
    {
        var total = order.Total;
        var lines = order.Items
            .Select(item => new OrderPlacedLine(item.ProductId, item.ProductName, item.Quantity, item.UnitPrice.Amount))
            .ToList();
        return new OrderPlaced(Guid.NewGuid(), now, order.Id, order.CustomerName, total.Amount, total.Currency, lines);
    }
}
