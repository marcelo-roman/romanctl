using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;

namespace SampleApp.Application.Orders.Queries;

public sealed record GetOrderById(Guid OrderId) : IQuery<OrderSummary>;

internal sealed class GetOrderByIdHandler(IReadModelStore<OrderSummary> store) : IQueryHandler<GetOrderById, OrderSummary>
{
    public async Task<OrderSummary> HandleAsync(GetOrderById query, CancellationToken cancellationToken) =>
        await store.FindAsync(query.OrderId, cancellationToken)
        ?? throw new NotFoundException($"Pedido {query.OrderId} não encontrado.");
}
