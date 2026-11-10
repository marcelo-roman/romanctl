using SampleApp.Ordering.Domain;

namespace SampleApp.Ordering.Application;

internal sealed record CreateOrder(string CustomerName);

internal sealed class CreateOrderHandler(IOrderRepository orders, TimeProvider clock)
{
    public async Task<Guid> HandleAsync(CreateOrder command, CancellationToken cancellationToken)
    {
        var order = Order.Create(command.CustomerName, clock.GetUtcNow());
        orders.Add(order);
        await orders.SaveChangesAsync(cancellationToken);
        return order.Id;
    }
}
