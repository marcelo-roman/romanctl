using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;
using {{.Name}}.Domain.Orders;

namespace {{.Name}}.Application.Orders.Commands;

public sealed record PlaceOrder(Guid OrderId) : ICommand<bool>;

internal sealed class PlaceOrderHandler(IOrderRepository orders, IUnitOfWork unitOfWork, TimeProvider clock)
    : ICommandHandler<PlaceOrder, bool>
{
    public async Task<bool> HandleAsync(PlaceOrder command, CancellationToken cancellationToken)
    {
        var order = await orders.GetByIdAsync(new OrderId(command.OrderId), cancellationToken)
            ?? throw new NotFoundException($"Pedido {command.OrderId} não encontrado.");

        order.Place(clock.GetUtcNow());
        await unitOfWork.SaveChangesAsync(cancellationToken);
        return true;
    }
}
