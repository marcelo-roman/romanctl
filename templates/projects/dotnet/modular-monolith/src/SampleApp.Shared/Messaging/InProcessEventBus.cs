using System.Threading.Channels;

namespace SampleApp.Shared.Messaging;

internal sealed class InProcessEventBus : IEventBus
{
    private readonly Channel<IIntegrationEvent> _channel = Channel.CreateUnbounded<IIntegrationEvent>(
        new UnboundedChannelOptions { SingleReader = true });

    public ValueTask PublishAsync(IIntegrationEvent integrationEvent, CancellationToken cancellationToken) =>
        _channel.Writer.WriteAsync(integrationEvent, cancellationToken);

    public IAsyncEnumerable<IIntegrationEvent> ReadAllAsync(CancellationToken cancellationToken) =>
        _channel.Reader.ReadAllAsync(cancellationToken);
}
