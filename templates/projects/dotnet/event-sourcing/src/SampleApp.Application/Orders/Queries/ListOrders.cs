using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;

namespace SampleApp.Application.Orders.Queries;

public sealed record ListOrders(int Limit = 50) : IQuery<IReadOnlyList<OrderSummary>>;

internal sealed class ListOrdersHandler(IReadModelStore<OrderSummary> store)
    : IQueryHandler<ListOrders, IReadOnlyList<OrderSummary>>
{
    public Task<IReadOnlyList<OrderSummary>> HandleAsync(ListOrders query, CancellationToken cancellationToken) =>
        store.ListAsync(Math.Clamp(query.Limit, 1, 200), cancellationToken);
}
