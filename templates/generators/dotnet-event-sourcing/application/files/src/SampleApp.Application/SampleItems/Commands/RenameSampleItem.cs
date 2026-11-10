using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Messaging;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Application.SampleItems.Commands;

public sealed record RenameSampleItem(Guid Id, string Name) : ICommand<bool>;

internal sealed class RenameSampleItemHandler(IAggregateRepository repository) : ICommandHandler<RenameSampleItem, bool>
{
    public async Task<bool> HandleAsync(RenameSampleItem command, CancellationToken cancellationToken)
    {
        var sampleItem = await repository.LoadAsync<SampleItem>(command.Id, cancellationToken)
            ?? throw new NotFoundException($"SampleItem {command.Id} não encontrado.");

        sampleItem.Rename(command.Name);
        await repository.SaveAsync(sampleItem, cancellationToken);
        return true;
    }
}
