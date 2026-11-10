using SampleApp.Domain.Common;

namespace SampleApp.Domain.SampleParents;

public sealed record SampleItem(SampleParentId SampleParentId, DateTimeOffset OccurredAt) : IDomainEvent;
