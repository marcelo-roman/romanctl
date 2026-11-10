using SampleApp.SampleItems.Domain;
using SampleApp.Shared.Domain;

namespace SampleApp.Tests.SampleItems;

public class SampleItemTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Criar_apara_o_nome()
    {
        var sampleItem = SampleItem.Create(" Exemplo ", Now);

        Assert.Equal("Exemplo", sampleItem.Name);
        Assert.Equal(Now, sampleItem.CreatedAt);
    }

    [Fact]
    public void Rejeita_nome_vazio() =>
        Assert.Throws<DomainException>(() => SampleItem.Create(" ", Now));

    [Fact]
    public void Renomear_rejeita_nome_vazio()
    {
        var sampleItem = SampleItem.Create("Exemplo", Now);

        Assert.Throws<DomainException>(() => sampleItem.Rename(""));
        Assert.Equal("Exemplo", sampleItem.Name);
    }
}
