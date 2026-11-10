namespace SampleApp.Application.EventSourcing;

public interface IReadModel
{
    Guid Id { get; }
    DateTimeOffset CreatedAt { get; }
}
