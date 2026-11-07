using {{.Name}}.Domain.Common;

namespace {{.Name}}.Domain.{{.InPlural}};

public sealed record {{.Entity}}({{.In}}Id {{.In}}Id, DateTimeOffset OccurredAt) : IDomainEvent;
