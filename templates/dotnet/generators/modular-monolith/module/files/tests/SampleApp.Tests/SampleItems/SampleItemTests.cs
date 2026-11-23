using SampleApp.SampleItems.Domain;
using SampleApp.Shared.Domain;

namespace SampleApp.Tests.SampleItems;

public class SampleItemTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Create_trims_the_name()
    {
        var sampleItem = SampleItem.Create(" Example ", Now);

        Assert.Equal("Example", sampleItem.Name);
        Assert.Equal(Now, sampleItem.CreatedAt);
    }

    [Fact]
    public void Rejects_empty_name() =>
        Assert.Throws<DomainException>(() => SampleItem.Create(" ", Now));

    [Fact]
    public void Rename_rejects_empty_name()
    {
        var sampleItem = SampleItem.Create("Example", Now);

        Assert.Throws<DomainException>(() => sampleItem.Rename(""));
        Assert.Equal("Example", sampleItem.Name);
    }
}
