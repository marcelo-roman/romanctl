using SampleApp.SampleParent.Domain;
using SampleApp.Shared.Application;

namespace SampleApp.SampleParent.Application;

internal sealed class GetSampleItemHandler(ISampleItemRepository sampleItems)
{
    public async Task<SampleItemResponse> HandleAsync(Guid id, CancellationToken cancellationToken)
    {
        var sampleItem = await sampleItems.FindAsync(id, cancellationToken)
            ?? throw new NotFoundException($"SampleItem {id} não encontrado.");
        return SampleItemResponse.From(sampleItem);
    }
}
