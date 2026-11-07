using {{.Name}}.Domain.Orders;

namespace {{.Name}}.Application.Orders;

public sealed record OrderItemDto(string ProductName, int Quantity, decimal UnitPrice, decimal Subtotal);

public sealed record OrderDto(
    Guid Id,
    string CustomerName,
    string Status,
    decimal Total,
    string Currency,
    DateTimeOffset CreatedAt,
    IReadOnlyList<OrderItemDto> Items)
{
    public static OrderDto From(Order order)
    {
        var total = order.Total;
        return new OrderDto(
            order.Id.Value,
            order.CustomerName,
            order.Status.ToString(),
            total.Amount,
            total.Currency,
            order.CreatedAt,
            order.Items
                .Select(item => new OrderItemDto(item.ProductName, item.Quantity, item.UnitPrice.Amount, item.Subtotal.Amount))
                .ToList());
    }
}
