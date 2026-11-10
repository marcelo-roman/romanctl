using JasperFx;
using JasperFx.Events.Projections;
using Marten;
using SampleApp.Application.EventSourcing;
using SampleApp.Domain.Common;

namespace SampleApp.Infrastructure.Persistence.Postgres;

internal static class MartenStoreOptions
{
    public static void UseDefaults(this StoreOptions options, string connectionString)
    {
        options.Connection(connectionString);
        options.AutoCreateSchemaObjects = AutoCreate.CreateOrUpdate;
        options.Events.AddEventTypes(DomainEventTypes());
    }

    public static void AddInlineProjection<TView>(this StoreOptions options, Projection<TView> projection)
        where TView : class, IReadModel =>
        options.Projections.Add(new MartenProjection<TView>(projection), ProjectionLifecycle.Inline);

    private static IEnumerable<Type> DomainEventTypes() =>
        typeof(IDomainEvent).Assembly.GetTypes()
            .Where(type => type is { IsClass: true, IsAbstract: false } && typeof(IDomainEvent).IsAssignableFrom(type));
}
