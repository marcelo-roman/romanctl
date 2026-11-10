using SampleApp.Ordering.Domain;
using SampleApp.Shared.Application;

namespace SampleApp.Ordering.Application;

internal sealed class GetOrderHandler(IOrderRepository orders)
{
    public async Task<OrderResponse> HandleAsync(Guid id, CancellationToken cancellationToken)
    {
        var order = await orders.FindAsync(id, cancellationToken)
            ?? throw new NotFoundException($"Pedido {id} não encontrado.");
        return OrderResponse.From(order);
    }
}
