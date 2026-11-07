using {{.Name}}.Application.Common;
using {{.Name}}.Application.Messaging;
using {{.Name}}.Domain.Common;
using {{.Name}}.Domain.Orders;
using Microsoft.EntityFrameworkCore;

namespace {{.Name}}.Infrastructure.Persistence.EfCore;

public sealed class AppDbContext(DbContextOptions<AppDbContext> options, IDomainEventPublisher publisher)
    : DbContext(options), IUnitOfWork
{
    public DbSet<Order> Orders => Set<Order>();

    protected override void OnModelCreating(ModelBuilder modelBuilder) =>
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(AppDbContext).Assembly);

    async Task IUnitOfWork.SaveChangesAsync(CancellationToken cancellationToken)
    {
        var aggregates = ChangeTracker.Entries<IAggregateRoot>()
            .Select(entry => entry.Entity)
            .Where(aggregate => aggregate.DomainEvents.Count > 0)
            .ToList();
        var events = aggregates.SelectMany(aggregate => aggregate.DomainEvents).ToList();

        await SaveChangesAsync(cancellationToken);

        aggregates.ForEach(aggregate => aggregate.ClearDomainEvents());
        foreach (var domainEvent in events)
            await publisher.PublishAsync(domainEvent, cancellationToken);
    }
}
