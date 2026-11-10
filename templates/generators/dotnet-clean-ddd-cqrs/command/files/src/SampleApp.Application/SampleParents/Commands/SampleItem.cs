using SampleApp.Application.Common;
using SampleApp.Application.Messaging;
using SampleApp.Domain.SampleParents;

namespace SampleApp.Application.SampleParents.Commands;

public sealed record SampleItem(Guid Id) : ICommand<bool>;

internal sealed class SampleItemHandler(ISampleParentRepository repository, IUnitOfWork unitOfWork)
    : ICommandHandler<SampleItem, bool>
{
    public async Task<bool> HandleAsync(SampleItem command, CancellationToken cancellationToken)
    {
        var sampleParent = await repository.GetByIdAsync(new SampleParentId(command.Id), cancellationToken)
            ?? throw new NotFoundException($"SampleParent {command.Id} não encontrado.");

        // Chame aqui o comportamento do agregado, por exemplo: sampleParent.Archive();

        await unitOfWork.SaveChangesAsync(cancellationToken);
        return true;
    }
}
