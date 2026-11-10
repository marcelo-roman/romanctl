using SampleApp.Ordering.Domain;
using SampleApp.Shared.Application;

namespace SampleApp.Ordering.Application;

internal sealed class ListOrdersHandler(IOrderRepository orders)
{
    public async Task<IReadOnlyList<OrderResponse>> HandleAsync(int? limit, CancellationToken cancellationToken)
    {
        var recent = await orders.ListRecentAsync(ListLimit.Clamp(limit), cancellationToken);
        return recent.Select(OrderResponse.From).ToList();
    }
}
