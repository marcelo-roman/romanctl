using {{.Name}}.Domain.Common;

namespace {{.Name}}.Domain.{{.InPlural}};

public sealed record {{.Entity}}
{
    public string Value { get; }

    public {{.Entity}}(string value)
    {
        if (string.IsNullOrWhiteSpace(value))
            throw new DomainException("{{.Entity}} não pode ser vazio.");

        Value = value.Trim();
    }

    public override string ToString() => Value;
}
