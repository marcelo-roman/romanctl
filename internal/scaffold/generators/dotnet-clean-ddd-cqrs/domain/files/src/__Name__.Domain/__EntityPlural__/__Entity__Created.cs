using {{.Name}}.Domain.Common;

namespace {{.Name}}.Domain.{{.EntityPlural}};

public sealed record {{.Entity}}Created({{.Entity}}Id {{.Entity}}Id, DateTimeOffset OccurredAt) : IDomainEvent;
