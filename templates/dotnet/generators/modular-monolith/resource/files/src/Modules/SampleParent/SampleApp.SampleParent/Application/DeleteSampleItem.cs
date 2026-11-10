using SampleApp.SampleParent.Domain;
using SampleApp.Shared.Application;

namespace SampleApp.SampleParent.Application;

internal sealed class DeleteSampleItemHandler(ISampleItemRepository sampleItems)
{
    public async Task HandleAsync(Guid id, CancellationToken cancellationToken)
    {
        var sampleItem = await sampleItems.FindAsync(id, cancellationToken)
            ?? throw new NotFoundException($"SampleItem {id} não encontrado.");

        sampleItems.Remove(sampleItem);
        await sampleItems.SaveChangesAsync(cancellationToken);
    }
}
