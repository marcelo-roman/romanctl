using {{.Name}}.Application.Messaging;
using {{.Name}}.Domain.{{.InPlural}};
using Microsoft.Extensions.Logging;

namespace {{.Name}}.Application.{{.InPlural}}.EventHandlers;

internal sealed class Log{{.Entity}}(ILogger<Log{{.Entity}}> logger) : IDomainEventHandler<{{.Entity}}>
{
    public Task HandleAsync({{.Entity}} domainEvent, CancellationToken cancellationToken)
    {
        logger.LogInformation("{{.Entity}}: {Id}", domainEvent.{{.In}}Id.Value);
        return Task.CompletedTask;
    }
}
