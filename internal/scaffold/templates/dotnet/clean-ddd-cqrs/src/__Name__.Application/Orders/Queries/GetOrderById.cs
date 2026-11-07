using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;

namespace {{.Name}}.Application.Orders.Queries;

public sealed record GetOrderById(Guid OrderId) : IQuery<OrderDto>;

internal sealed class GetOrderByIdHandler(IOrderReadStore store) : IQueryHandler<GetOrderById, OrderDto>
{
    public async Task<OrderDto> HandleAsync(GetOrderById query, CancellationToken cancellationToken) =>
        await store.FindAsync(query.OrderId, cancellationToken)
        ?? throw new NotFoundException($"Pedido {query.OrderId} não encontrado.");
}
