using SampleApp.Domain.Common;

namespace SampleApp.Domain.Orders;

public sealed record OrderCreated(string CustomerName, DateTimeOffset CreatedAt) : IDomainEvent;
