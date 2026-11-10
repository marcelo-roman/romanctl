using SampleApp.SampleItems.Domain;

namespace SampleApp.SampleItems.Application;

internal sealed record SampleItemResponse(Guid Id, string Name, DateTimeOffset CreatedAt)
{
    public static SampleItemResponse From(SampleItem sampleItem) => new(sampleItem.Id, sampleItem.Name, sampleItem.CreatedAt);
}
