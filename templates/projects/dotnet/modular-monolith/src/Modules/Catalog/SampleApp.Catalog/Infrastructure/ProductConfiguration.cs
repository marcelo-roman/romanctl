using SampleApp.Catalog.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace SampleApp.Catalog.Infrastructure;

internal sealed class ProductConfiguration : IEntityTypeConfiguration<Product>
{
    public void Configure(EntityTypeBuilder<Product> builder)
    {
        builder.ToTable("products");
        builder.HasKey(product => product.Id);
        builder.Property(product => product.Id).ValueGeneratedNever();
        builder.Property(product => product.Name).HasMaxLength(200).IsRequired();
        builder.OwnsOne(product => product.Price, price =>
        {
            price.Property(money => money.Amount).HasColumnName("price").HasPrecision(18, 2);
            price.Property(money => money.Currency).HasColumnName("currency").HasMaxLength(3);
        });
        builder.Navigation(product => product.Price).IsRequired();
    }
}
