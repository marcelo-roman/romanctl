using SampleApp.Domain.Common;

namespace SampleApp.Domain.Orders;

public sealed record ItemAdded(string ProductName, int Quantity, decimal UnitPrice, string Currency) : IDomainEvent;
