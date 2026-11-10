using SampleApp.Application.EventSourcing;
using SampleApp.Domain.Orders;

namespace SampleApp.Application.Orders;

public sealed class OrderSummaryProjection : Projection<OrderSummary>
{
    public OrderSummaryProjection()
    {
        Creates<OrderCreated>((id, created) => new OrderSummary
        {
            Id = id,
            CustomerName = created.CustomerName,
            Status = nameof(OrderStatus.Draft),
            Currency = Order.DefaultCurrency,
            CreatedAt = created.CreatedAt,
        });

        Updates<ItemAdded>((summary, added) =>
        {
            var subtotal = added.UnitPrice * added.Quantity;
            return summary with
            {
                Currency = added.Currency,
                Total = summary.Total + subtotal,
                Items = [.. summary.Items, new OrderSummaryItem(added.ProductName, added.Quantity, added.UnitPrice, subtotal)],
            };
        });

        Updates<OrderPlaced>((summary, placed) => summary with
        {
            Status = nameof(OrderStatus.Placed),
            Total = placed.Total,
            Currency = placed.Currency,
            PlacedAt = placed.PlacedAt,
        });
    }
}
