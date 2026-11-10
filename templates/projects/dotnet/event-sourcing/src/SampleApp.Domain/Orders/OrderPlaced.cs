using SampleApp.Domain.Common;

namespace SampleApp.Domain.Orders;

public sealed record OrderPlaced(decimal Total, string Currency, DateTimeOffset PlacedAt) : IDomainEvent;
