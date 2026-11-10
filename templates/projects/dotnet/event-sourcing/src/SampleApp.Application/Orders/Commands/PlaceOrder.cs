using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;
using SampleApp.Domain.Orders;

namespace SampleApp.Application.Orders.Commands;

public sealed record PlaceOrder(Guid OrderId) : ICommand<bool>;

internal sealed class PlaceOrderHandler(IAggregateRepository repository, TimeProvider clock) : ICommandHandler<PlaceOrder, bool>
{
    public async Task<bool> HandleAsync(PlaceOrder command, CancellationToken cancellationToken)
    {
        var order = await repository.LoadAsync<Order>(command.OrderId, cancellationToken)
            ?? throw new NotFoundException($"Pedido {command.OrderId} não encontrado.");

        order.Place(clock.GetUtcNow());
        await repository.SaveAsync(order, cancellationToken);
        return true;
    }
}
