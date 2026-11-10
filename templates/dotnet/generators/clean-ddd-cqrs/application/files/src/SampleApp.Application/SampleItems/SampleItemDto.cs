using SampleApp.Domain.SampleItems;

namespace SampleApp.Application.SampleItems;

public sealed record SampleItemDto(Guid Id, string Name, DateTimeOffset CreatedAt)
{
    public static SampleItemDto From(SampleItem sampleItem) =>
        new(sampleItem.Id.Value, sampleItem.Name, sampleItem.CreatedAt);
}
