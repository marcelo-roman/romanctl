using SampleApp.Application.Common;
using SampleApp.Application.Messaging;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Application.SampleItems.Commands;

public sealed record CreateSampleItem(string Name) : ICommand<Guid>;

internal sealed class CreateSampleItemHandler(ISampleItemRepository repository, IUnitOfWork unitOfWork, TimeProvider clock)
    : ICommandHandler<CreateSampleItem, Guid>
{
    public async Task<Guid> HandleAsync(CreateSampleItem command, CancellationToken cancellationToken)
    {
        var sampleItem = SampleItem.Create(command.Name, clock.GetUtcNow());
        repository.Add(sampleItem);
        await unitOfWork.SaveChangesAsync(cancellationToken);
        return sampleItem.Id.Value;
    }
}
