using SampleApp.Application.EventSourcing;

namespace SampleApp.Application.SampleItems;

public sealed record SampleItemSummary : IReadModel
{
    public Guid Id { get; init; }
    public string Name { get; init; } = "";
    public DateTimeOffset CreatedAt { get; init; }
    public int Renames { get; init; }
}
