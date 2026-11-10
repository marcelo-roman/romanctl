using SampleApp.Application.Messaging;
using SampleApp.Domain.SampleParents;
using Microsoft.Extensions.Logging;

namespace SampleApp.Application.SampleParents.EventHandlers;

internal sealed class LogSampleItem(ILogger<LogSampleItem> logger) : IDomainEventHandler<SampleItem>
{
    public Task HandleAsync(SampleItem domainEvent, CancellationToken cancellationToken)
    {
        logger.LogInformation("SampleItem: {Id}", domainEvent.SampleParentId.Value);
        return Task.CompletedTask;
    }
}
