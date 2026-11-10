using SampleApp.SampleParent.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace SampleApp.SampleParent.Infrastructure;

internal sealed class SampleItemConfiguration : IEntityTypeConfiguration<SampleItem>
{
    public void Configure(EntityTypeBuilder<SampleItem> builder)
    {
        builder.ToTable("sample_items");
        builder.HasKey(sampleItem => sampleItem.Id);
        builder.Property(sampleItem => sampleItem.Id).ValueGeneratedNever();
        builder.Property(sampleItem => sampleItem.Name).HasMaxLength(200).IsRequired();
    }
}
