using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;
using SampleApp.Domain.Orders;

namespace SampleApp.Application.Orders.Commands;

public sealed record CreateOrder(string CustomerName) : ICommand<Guid>;

internal sealed class CreateOrderHandler(IAggregateRepository repository, TimeProvider clock)
    : ICommandHandler<CreateOrder, Guid>
{
    public async Task<Guid> HandleAsync(CreateOrder command, CancellationToken cancellationToken)
    {
        var order = Order.Create(Guid.NewGuid(), command.CustomerName, clock.GetUtcNow());
        await repository.SaveAsync(order, cancellationToken);
        return order.Id;
    }
}
