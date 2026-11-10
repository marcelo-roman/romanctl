using SampleApp.SampleItems.Domain;
using SampleApp.Shared.Application;

namespace SampleApp.SampleItems.Application;

internal sealed class ListSampleItemsHandler(ISampleItemRepository sampleItems)
{
    public async Task<IReadOnlyList<SampleItemResponse>> HandleAsync(int? limit, CancellationToken cancellationToken)
    {
        var recent = await sampleItems.ListRecentAsync(ListLimit.Clamp(limit), cancellationToken);
        return recent.Select(SampleItemResponse.From).ToList();
    }
}
