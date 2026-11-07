using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;
using {{.Name}}.Domain.Orders;

namespace {{.Name}}.Application.Orders.Commands;

public sealed record CreateOrder(string CustomerName) : ICommand<Guid>;

internal sealed class CreateOrderHandler(IOrderRepository orders, IUnitOfWork unitOfWork, TimeProvider clock)
    : ICommandHandler<CreateOrder, Guid>
{
    public async Task<Guid> HandleAsync(CreateOrder command, CancellationToken cancellationToken)
    {
        var order = Order.Create(command.CustomerName, clock.GetUtcNow());
        orders.Add(order);
        await unitOfWork.SaveChangesAsync(cancellationToken);
        return order.Id.Value;
    }
}
