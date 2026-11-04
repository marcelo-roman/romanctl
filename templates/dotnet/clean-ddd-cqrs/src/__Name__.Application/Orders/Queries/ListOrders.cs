using {{.Name}}.Application.Messaging;

namespace {{.Name}}.Application.Orders.Queries;

public sealed record ListOrders(int Limit = 50) : IQuery<IReadOnlyList<OrderDto>>;

internal sealed class ListOrdersHandler(IOrderReadStore store) : IQueryHandler<ListOrders, IReadOnlyList<OrderDto>>
{
    public Task<IReadOnlyList<OrderDto>> HandleAsync(ListOrders query, CancellationToken cancellationToken) =>
        store.ListAsync(Math.Clamp(query.Limit, 1, 200), cancellationToken);
}
