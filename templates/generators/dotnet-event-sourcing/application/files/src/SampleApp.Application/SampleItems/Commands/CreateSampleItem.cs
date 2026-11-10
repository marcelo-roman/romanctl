using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Application.SampleItems.Commands;

public sealed record CreateSampleItem(string Name) : ICommand<Guid>;

internal sealed class CreateSampleItemHandler(IAggregateRepository repository, TimeProvider clock)
    : ICommandHandler<CreateSampleItem, Guid>
{
    public async Task<Guid> HandleAsync(CreateSampleItem command, CancellationToken cancellationToken)
    {
        var sampleItem = SampleItem.Create(Guid.NewGuid(), command.Name, clock.GetUtcNow());
        await repository.SaveAsync(sampleItem, cancellationToken);
        return sampleItem.Id;
    }
}
