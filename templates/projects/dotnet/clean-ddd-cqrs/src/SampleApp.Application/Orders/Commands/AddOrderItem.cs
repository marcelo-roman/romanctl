using SampleApp.Application.Common;
using SampleApp.Application.Messaging;
using SampleApp.Domain.Orders;

namespace SampleApp.Application.Orders.Commands;

public sealed record AddOrderItem(Guid OrderId, string ProductName, int Quantity, decimal UnitPrice, string? Currency = null)
    : ICommand<bool>;

internal sealed class AddOrderItemHandler(IOrderRepository orders, IUnitOfWork unitOfWork)
    : ICommandHandler<AddOrderItem, bool>
{
    public async Task<bool> HandleAsync(AddOrderItem command, CancellationToken cancellationToken)
    {
        var order = await orders.GetByIdAsync(new OrderId(command.OrderId), cancellationToken)
            ?? throw new NotFoundException($"Pedido {command.OrderId} não encontrado.");

        order.AddItem(command.ProductName, command.Quantity, new Money(command.UnitPrice, command.Currency ?? Order.DefaultCurrency));
        await unitOfWork.SaveChangesAsync(cancellationToken);
        return true;
    }
}
