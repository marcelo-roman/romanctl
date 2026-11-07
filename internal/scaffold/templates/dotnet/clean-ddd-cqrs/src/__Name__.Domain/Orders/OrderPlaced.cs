using {{.Name}}.Domain.Common;

namespace {{.Name}}.Domain.Orders;

public sealed record OrderPlaced(OrderId OrderId, decimal Total, string Currency, DateTimeOffset OccurredAt) : IDomainEvent;
