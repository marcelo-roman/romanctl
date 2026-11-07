using {{.Name}}.Domain.{{.EntityPlural}};
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace {{.Name}}.Infrastructure.Persistence.EfCore;

internal sealed class {{.Entity}}Configuration : IEntityTypeConfiguration<{{.Entity}}>
{
    public void Configure(EntityTypeBuilder<{{.Entity}}> builder)
    {
        builder.ToTable("{{.EntityPluralSnake}}");
        builder.HasKey({{.EntityCamel}} => {{.EntityCamel}}.Id);
        builder.Property({{.EntityCamel}} => {{.EntityCamel}}.Id)
            .HasConversion(id => id.Value, value => new {{.Entity}}Id(value))
            .ValueGeneratedNever();
        builder.Property({{.EntityCamel}} => {{.EntityCamel}}.Name).HasMaxLength(200).IsRequired();
        builder.Ignore({{.EntityCamel}} => {{.EntityCamel}}.DomainEvents);
    }
}
