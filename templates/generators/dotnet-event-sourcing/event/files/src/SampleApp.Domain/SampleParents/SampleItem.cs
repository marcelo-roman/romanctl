using SampleApp.Domain.Common;

namespace SampleApp.Domain.SampleParents;

public sealed record SampleItem(DateTimeOffset OccurredAt) : IDomainEvent;
