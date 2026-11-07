using {{.Name}}.Domain.Common;

namespace {{.Name}}.Domain.{{.EntityPlural}};

public sealed class {{.Entity}} : AggregateRoot<{{.Entity}}Id>
{
    public string Name { get; private set; } = default!;
    public DateTimeOffset CreatedAt { get; private set; }

    private {{.Entity}}() { }

    public static {{.Entity}} Create(string name, DateTimeOffset now)
    {
        var {{.EntityCamel}} = new {{.Entity}} { Id = {{.Entity}}Id.New(), CreatedAt = now };
        {{.EntityCamel}}.Rename(name);
        {{.EntityCamel}}.Raise(new {{.Entity}}Created({{.EntityCamel}}.Id, now));
        return {{.EntityCamel}};
    }

    public void Rename(string name)
    {
        if (string.IsNullOrWhiteSpace(name))
            throw new DomainException("O nome é obrigatório.");

        Name = name.Trim();
    }
}
