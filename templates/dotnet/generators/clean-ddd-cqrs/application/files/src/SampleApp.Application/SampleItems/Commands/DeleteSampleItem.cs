using SampleApp.Application.Common;
using SampleApp.Application.Messaging;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Application.SampleItems.Commands;

public sealed record DeleteSampleItem(Guid Id) : ICommand<bool>;

internal sealed class DeleteSampleItemHandler(ISampleItemRepository repository, IUnitOfWork unitOfWork)
    : ICommandHandler<DeleteSampleItem, bool>
{
    public async Task<bool> HandleAsync(DeleteSampleItem command, CancellationToken cancellationToken)
    {
        var sampleItem = await repository.GetByIdAsync(new SampleItemId(command.Id), cancellationToken)
            ?? throw new NotFoundException($"SampleItem {command.Id} não encontrado.");

        repository.Remove(sampleItem);
        await unitOfWork.SaveChangesAsync(cancellationToken);
        return true;
    }
}
