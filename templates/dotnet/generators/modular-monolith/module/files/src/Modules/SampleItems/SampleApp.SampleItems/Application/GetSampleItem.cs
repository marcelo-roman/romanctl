using SampleApp.SampleItems.Domain;
using SampleApp.Shared.Application;

namespace SampleApp.SampleItems.Application;

internal sealed class GetSampleItemHandler(ISampleItemRepository sampleItems)
{
    public async Task<SampleItemResponse> HandleAsync(Guid id, CancellationToken cancellationToken)
    {
        var sampleItem = await sampleItems.FindAsync(id, cancellationToken)
            ?? throw new NotFoundException($"SampleItem {id} not found.");
        return SampleItemResponse.From(sampleItem);
    }
}
