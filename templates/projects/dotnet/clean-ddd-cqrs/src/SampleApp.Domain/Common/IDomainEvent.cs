namespace SampleApp.Domain.Common;

public interface IDomainEvent
{
    DateTimeOffset OccurredAt { get; }
}
