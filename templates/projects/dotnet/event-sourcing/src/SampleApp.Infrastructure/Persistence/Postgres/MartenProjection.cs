using JasperFx.Events;
using Marten.Events.Aggregation;
using SampleApp.Application.EventSourcing;
using SampleApp.Domain.Common;

namespace SampleApp.Infrastructure.Persistence.Postgres;

internal sealed class MartenProjection<TView> : SingleStreamProjection<TView, Guid> where TView : class, IReadModel
{
    private readonly Projection<TView> _projection;

    public MartenProjection(Projection<TView> projection)
    {
        _projection = projection;
        foreach (var eventType in projection.EventTypes)
            IncludeType(eventType);
    }

    public override TView? Evolve(TView? snapshot, Guid id, IEvent e) =>
        e.Data is IDomainEvent domainEvent ? _projection.Evolve(snapshot, id, domainEvent) : snapshot;
}
