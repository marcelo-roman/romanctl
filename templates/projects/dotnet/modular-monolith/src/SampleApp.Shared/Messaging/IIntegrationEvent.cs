namespace SampleApp.Shared.Messaging;

public interface IIntegrationEvent
{
    Guid EventId { get; }
    DateTimeOffset OccurredAt { get; }
}
