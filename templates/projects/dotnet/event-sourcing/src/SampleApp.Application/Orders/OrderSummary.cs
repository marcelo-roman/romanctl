using SampleApp.Application.EventSourcing;

namespace SampleApp.Application.Orders;

public sealed record OrderSummaryItem(string ProductName, int Quantity, decimal UnitPrice, decimal Subtotal);

public sealed record OrderSummary : IReadModel
{
    public Guid Id { get; init; }
    public string CustomerName { get; init; } = "";
    public string Status { get; init; } = "";
    public decimal Total { get; init; }
    public string Currency { get; init; } = "";
    public DateTimeOffset CreatedAt { get; init; }
    public DateTimeOffset? PlacedAt { get; init; }
    public IReadOnlyList<OrderSummaryItem> Items { get; init; } = [];
}
