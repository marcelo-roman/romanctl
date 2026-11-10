using SampleApp.Domain.Common;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Tests;

public class SampleItemTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Criar_levanta_evento()
    {
        var sampleItem = SampleItem.Create(" Exemplo ", Now);

        Assert.Equal("Exemplo", sampleItem.Name);
        Assert.IsType<SampleItemCreated>(Assert.Single(sampleItem.DomainEvents));
    }

    [Fact]
    public void Rejeita_nome_vazio() =>
        Assert.Throws<DomainException>(() => SampleItem.Create(" ", Now));
}
