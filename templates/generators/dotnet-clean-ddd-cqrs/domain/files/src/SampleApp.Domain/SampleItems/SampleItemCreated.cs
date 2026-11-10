using SampleApp.Domain.Common;

namespace SampleApp.Domain.SampleItems;

public sealed record SampleItemCreated(SampleItemId SampleItemId, DateTimeOffset OccurredAt) : IDomainEvent;
