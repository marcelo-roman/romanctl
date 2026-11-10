using SampleApp.SampleItems.Domain;
using SampleApp.Shared.Application;

namespace SampleApp.SampleItems.Application;

internal sealed record RenameSampleItem(Guid Id, string Name);

internal sealed class RenameSampleItemHandler(ISampleItemRepository sampleItems)
{
    public async Task HandleAsync(RenameSampleItem command, CancellationToken cancellationToken)
    {
        var sampleItem = await sampleItems.FindAsync(command.Id, cancellationToken)
            ?? throw new NotFoundException($"SampleItem {command.Id} não encontrado.");

        sampleItem.Rename(command.Name);
        await sampleItems.SaveChangesAsync(cancellationToken);
    }
}
