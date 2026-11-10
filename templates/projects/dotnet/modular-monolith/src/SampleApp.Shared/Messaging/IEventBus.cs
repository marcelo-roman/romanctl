namespace SampleApp.Shared.Messaging;

public interface IEventBus
{
    ValueTask PublishAsync(IIntegrationEvent integrationEvent, CancellationToken cancellationToken);
}
