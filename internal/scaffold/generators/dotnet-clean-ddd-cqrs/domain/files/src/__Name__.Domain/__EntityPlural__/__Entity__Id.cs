namespace {{.Name}}.Domain.{{.EntityPlural}};

public readonly record struct {{.Entity}}Id(Guid Value)
{
    public static {{.Entity}}Id New() => new(Guid.NewGuid());
}
