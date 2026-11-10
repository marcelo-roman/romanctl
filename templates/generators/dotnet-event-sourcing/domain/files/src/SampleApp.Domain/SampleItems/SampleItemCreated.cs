using SampleApp.Domain.Common;

namespace SampleApp.Domain.SampleItems;

public sealed record SampleItemCreated(string Name, DateTimeOffset CreatedAt) : IDomainEvent;
