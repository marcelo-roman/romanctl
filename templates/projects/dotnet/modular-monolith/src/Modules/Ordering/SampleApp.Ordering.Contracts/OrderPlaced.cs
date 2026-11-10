using SampleApp.Shared.Messaging;

namespace SampleApp.Ordering.Contracts;

public sealed record OrderPlaced(
    Guid EventId,
    DateTimeOffset OccurredAt,
    Guid OrderId,
    string CustomerName,
    decimal Total,
    string Currency,
    IReadOnlyList<OrderPlacedLine> Lines) : IIntegrationEvent;
