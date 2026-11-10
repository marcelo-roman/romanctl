using SampleApp.Application.Common;
using SampleApp.Application.Messaging;
using SampleApp.Domain.Orders;

namespace SampleApp.Application.Orders.Commands;

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
