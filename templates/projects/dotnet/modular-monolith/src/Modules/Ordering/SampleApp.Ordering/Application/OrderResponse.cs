using SampleApp.Ordering.Domain;

namespace SampleApp.Ordering.Application;

internal sealed record OrderItemResponse(Guid ProductId, string ProductName, int Quantity, decimal UnitPrice, decimal Subtotal);

internal sealed record OrderResponse(
    Guid Id,
    string CustomerName,
    string Status,
    decimal Total,
    string Currency,
    DateTimeOffset CreatedAt,
    IReadOnlyList<OrderItemResponse> Items)
{
    public static OrderResponse From(Order order)
    {
        var total = order.Total;
        return new OrderResponse(
            order.Id,
            order.CustomerName,
            order.Status.ToString(),
            total.Amount,
            total.Currency,
            order.CreatedAt,
            order.Items
                .Select(item => new OrderItemResponse(item.ProductId, item.ProductName, item.Quantity, item.UnitPrice.Amount, item.Subtotal.Amount))
                .ToList());
    }
}
