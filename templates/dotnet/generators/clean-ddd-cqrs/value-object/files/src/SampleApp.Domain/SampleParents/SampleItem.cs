using SampleApp.Domain.Common;

namespace SampleApp.Domain.SampleParents;

public sealed record SampleItem
{
    public string Value { get; }

    public SampleItem(string value)
    {
        if (string.IsNullOrWhiteSpace(value))
            throw new DomainException("SampleItem não pode ser vazio.");

        Value = value.Trim();
    }

    public override string ToString() => Value;
}
