using SampleApp.Domain.SampleItems;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace SampleApp.Infrastructure.Persistence.EfCore;

internal sealed class SampleItemConfiguration : IEntityTypeConfiguration<SampleItem>
{
    public void Configure(EntityTypeBuilder<SampleItem> builder)
    {
        builder.ToTable("sample_items");
        builder.HasKey(sampleItem => sampleItem.Id);
        builder.Property(sampleItem => sampleItem.Id)
            .HasConversion(id => id.Value, value => new SampleItemId(value))
            .ValueGeneratedNever();
        builder.Property(sampleItem => sampleItem.Name).HasMaxLength(200).IsRequired();
        builder.Ignore(sampleItem => sampleItem.DomainEvents);
    }
}
