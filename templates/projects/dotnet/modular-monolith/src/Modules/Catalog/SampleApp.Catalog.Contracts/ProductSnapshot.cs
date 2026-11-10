namespace SampleApp.Catalog.Contracts;

public sealed record ProductSnapshot(Guid Id, string Name, decimal Price, string Currency);
