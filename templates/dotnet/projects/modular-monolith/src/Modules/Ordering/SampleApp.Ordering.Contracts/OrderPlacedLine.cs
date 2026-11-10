namespace SampleApp.Ordering.Contracts;

public sealed record OrderPlacedLine(Guid ProductId, string ProductName, int Quantity, decimal UnitPrice);
