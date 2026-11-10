using SampleApp.Domain.Common;

namespace SampleApp.Application.EventSourcing;

public sealed record StoredEvent(Guid StreamId, long Version, IDomainEvent Data, DateTimeOffset RecordedAt);
