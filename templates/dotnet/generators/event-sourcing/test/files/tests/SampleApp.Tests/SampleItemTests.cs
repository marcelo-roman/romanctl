using SampleApp.Application.SampleItems;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Tests;

public class SampleItemTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);
    private static readonly SampleItemCreated Created = new("First", Now);

    private static AggregateScenario<SampleItem> Scenario() => new();

    [Fact]
    public void Create_raises_SampleItemCreated() =>
        Scenario()
            .WhenCreating(id => SampleItem.Create(id, " First ", Now))
            .Then(Created);

    [Fact]
    public void Create_requires_name() =>
        Scenario()
            .WhenCreating(id => SampleItem.Create(id, " ", Now))
            .ThenFails("Name is required.");

    [Fact]
    public void Rename_raises_SampleItemRenamed() =>
        Scenario()
            .Given(Created)
            .When(sampleItem => sampleItem.Rename("Second"))
            .Then(new SampleItemRenamed("Second"));

    [Fact]
    public void Rename_to_the_same_name_raises_no_event() =>
        Scenario()
            .Given(Created)
            .When(sampleItem => sampleItem.Rename(" First "))
            .Then();

    [Fact]
    public void Rename_requires_name() =>
        Scenario()
            .Given(Created)
            .When(sampleItem => sampleItem.Rename(""))
            .ThenFails("Name is required.");

    [Fact]
    public void Projection_follows_the_events()
    {
        var id = Guid.NewGuid();
        var projection = new SampleItemSummaryProjection();

        var summary = projection.Evolve(projection.Evolve(null, id, Created), id, new SampleItemRenamed("Second"));

        Assert.Equal(new SampleItemSummary { Id = id, Name = "Second", CreatedAt = Now, Renames = 1 }, summary);
    }
}
