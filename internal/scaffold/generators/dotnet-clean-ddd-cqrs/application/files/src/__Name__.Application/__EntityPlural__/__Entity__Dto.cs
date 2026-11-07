using {{.Name}}.Domain.{{.EntityPlural}};

namespace {{.Name}}.Application.{{.EntityPlural}};

public sealed record {{.Entity}}Dto(Guid Id, string Name, DateTimeOffset CreatedAt)
{
    public static {{.Entity}}Dto From({{.Entity}} {{.EntityCamel}}) =>
        new({{.EntityCamel}}.Id.Value, {{.EntityCamel}}.Name, {{.EntityCamel}}.CreatedAt);
}
