using SampleApp.Domain.Common;

namespace SampleApp.Domain.SampleItems;

public sealed record SampleItemRenamed(string Name) : IDomainEvent;
