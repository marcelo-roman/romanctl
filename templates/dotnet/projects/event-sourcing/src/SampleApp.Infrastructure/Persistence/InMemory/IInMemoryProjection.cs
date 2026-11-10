using SampleApp.Application.EventSourcing;

namespace SampleApp.Infrastructure.Persistence.InMemory;

internal interface IInMemoryProjection
{
    void Apply(StoredEvent stored);
}
