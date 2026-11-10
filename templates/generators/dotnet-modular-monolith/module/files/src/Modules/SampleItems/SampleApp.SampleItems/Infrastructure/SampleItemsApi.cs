using SampleApp.SampleItems.Contracts;
using SampleApp.SampleItems.Domain;

namespace SampleApp.SampleItems.Infrastructure;

internal sealed class SampleItemsApi(ISampleItemRepository sampleItems) : ISampleItemsApi
{
    public async Task<SampleItemSnapshot?> FindSampleItemAsync(Guid id, CancellationToken cancellationToken)
    {
        var sampleItem = await sampleItems.FindAsync(id, cancellationToken);
        if (sampleItem is null)
            return null;

        return new SampleItemSnapshot(sampleItem.Id, sampleItem.Name);
    }
}
