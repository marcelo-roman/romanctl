using SampleApp.Shared.Messaging;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

namespace SampleApp.Tests.Shared;

public class EventBusTests
{
    public sealed record SomethingHappened(Guid EventId, DateTimeOffset OccurredAt) : IIntegrationEvent;

    public sealed class Received
    {
        public TaskCompletionSource<SomethingHappened> Event { get; } = new(TaskCreationOptions.RunContinuationsAsynchronously);
    }

    public sealed class RecordingHandler(Received received) : IIntegrationEventHandler<SomethingHappened>
    {
        public Task HandleAsync(SomethingHappened integrationEvent, CancellationToken cancellationToken)
        {
            received.Event.TrySetResult(integrationEvent);
            return Task.CompletedTask;
        }
    }

    public sealed class FailingHandler : IIntegrationEventHandler<SomethingHappened>
    {
        public Task HandleAsync(SomethingHappened integrationEvent, CancellationToken cancellationToken) =>
            throw new InvalidOperationException("Deliberate failure.");
    }

    [Fact]
    public async Task Delivers_the_event_to_handlers_outside_the_request_even_if_one_fails()
    {
        var received = new Received();
        await using var provider = new ServiceCollection()
            .AddLogging()
            .AddSingleton(received)
            .AddEventBus()
            .AddIntegrationEventHandler<SomethingHappened, FailingHandler>()
            .AddIntegrationEventHandler<SomethingHappened, RecordingHandler>()
            .BuildServiceProvider();
        var dispatcher = provider.GetServices<IHostedService>().Single();
        await dispatcher.StartAsync(CancellationToken.None);

        var published = new SomethingHappened(Guid.NewGuid(), DateTimeOffset.UtcNow);
        await provider.GetRequiredService<IEventBus>().PublishAsync(published, CancellationToken.None);

        Assert.Equal(published, await received.Event.Task.WaitAsync(TimeSpan.FromSeconds(10)));
        await dispatcher.StopAsync(CancellationToken.None);
    }
}
