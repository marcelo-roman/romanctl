using SampleApp.Ordering.Domain;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace SampleApp.Ordering.Infrastructure;

internal sealed class OrderConfiguration : IEntityTypeConfiguration<Order>
{
    public void Configure(EntityTypeBuilder<Order> builder)
    {
        builder.ToTable("orders");
        builder.HasKey(order => order.Id);
        builder.Property(order => order.Id).ValueGeneratedNever();
        builder.Property(order => order.CustomerName).HasMaxLength(200).IsRequired();
        builder.Property(order => order.Status).HasConversion<string>().HasMaxLength(20);
        builder.Ignore(order => order.Total);

        builder.OwnsMany(order => order.Items, items =>
        {
            items.ToTable("order_items");
            items.WithOwner().HasForeignKey("OrderId");
            items.HasKey(item => item.Id);
            items.Property(item => item.Id).ValueGeneratedNever();
            items.Property(item => item.ProductName).HasMaxLength(200).IsRequired();
            items.Ignore(item => item.Subtotal);
            items.OwnsOne(item => item.UnitPrice, price =>
            {
                price.Property(money => money.Amount).HasColumnName("unit_price").HasPrecision(18, 2);
                price.Property(money => money.Currency).HasColumnName("currency").HasMaxLength(3);
            });
        });
        builder.Navigation(order => order.Items).UsePropertyAccessMode(PropertyAccessMode.Field);
    }
}
