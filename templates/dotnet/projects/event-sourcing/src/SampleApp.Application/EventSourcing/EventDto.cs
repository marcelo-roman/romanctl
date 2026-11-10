namespace SampleApp.Application.EventSourcing;

public sealed record EventDto(long Version, string Type, DateTimeOffset RecordedAt, object Data)
{
    public static EventDto From(StoredEvent stored) =>
        new(stored.Version, stored.Data.GetType().Name, stored.RecordedAt, stored.Data);
}
