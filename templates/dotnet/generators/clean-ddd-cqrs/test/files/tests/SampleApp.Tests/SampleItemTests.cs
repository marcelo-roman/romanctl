using SampleApp.Domain.Common;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Tests;

public class SampleItemTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Create_raises_event()
    {
        var sampleItem = SampleItem.Create(" Example ", Now);

        Assert.Equal("Example", sampleItem.Name);
        Assert.IsType<SampleItemCreated>(Assert.Single(sampleItem.DomainEvents));
    }

    [Fact]
    public void Rejects_empty_name() =>
        Assert.Throws<DomainException>(() => SampleItem.Create(" ", Now));
}
