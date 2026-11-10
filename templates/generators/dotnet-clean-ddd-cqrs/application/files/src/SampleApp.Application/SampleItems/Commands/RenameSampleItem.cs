using SampleApp.Application.Common;
using SampleApp.Application.Messaging;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Application.SampleItems.Commands;

public sealed record RenameSampleItem(Guid Id, string Name) : ICommand<bool>;

internal sealed class RenameSampleItemHandler(ISampleItemRepository repository, IUnitOfWork unitOfWork)
    : ICommandHandler<RenameSampleItem, bool>
{
    public async Task<bool> HandleAsync(RenameSampleItem command, CancellationToken cancellationToken)
    {
        var sampleItem = await repository.GetByIdAsync(new SampleItemId(command.Id), cancellationToken)
            ?? throw new NotFoundException($"SampleItem {command.Id} não encontrado.");

        sampleItem.Rename(command.Name);
        await unitOfWork.SaveChangesAsync(cancellationToken);
        return true;
    }
}
