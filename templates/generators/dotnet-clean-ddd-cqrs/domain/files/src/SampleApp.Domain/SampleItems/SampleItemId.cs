namespace SampleApp.Domain.SampleItems;

public readonly record struct SampleItemId(Guid Value)
{
    public static SampleItemId New() => new(Guid.NewGuid());
}
