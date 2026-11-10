using SampleApp.Domain.Common;

namespace SampleApp.Domain.Orders;

public sealed record OrderPlaced(OrderId OrderId, decimal Total, string Currency, DateTimeOffset OccurredAt) : IDomainEvent;
