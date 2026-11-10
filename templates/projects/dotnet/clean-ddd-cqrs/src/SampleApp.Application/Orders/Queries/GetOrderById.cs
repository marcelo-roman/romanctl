using SampleApp.Application.Common;
using SampleApp.Application.Messaging;

namespace SampleApp.Application.Orders.Queries;

public sealed record GetOrderById(Guid OrderId) : IQuery<OrderDto>;

internal sealed class GetOrderByIdHandler(IOrderReadStore store) : IQueryHandler<GetOrderById, OrderDto>
{
    public async Task<OrderDto> HandleAsync(GetOrderById query, CancellationToken cancellationToken) =>
        await store.FindAsync(query.OrderId, cancellationToken)
        ?? throw new NotFoundException($"Pedido {query.OrderId} não encontrado.");
}
