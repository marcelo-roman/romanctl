using SampleApp.Application.SampleItems;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Tests;

public class SampleItemTests
{
    private static readonly DateTimeOffset Now = new(2026, 1, 1, 12, 0, 0, TimeSpan.Zero);
    private static readonly SampleItemCreated Created = new("Primeiro", Now);

    private static AggregateScenario<SampleItem> Scenario() => new();

    [Fact]
    public void Criar_gera_SampleItemCreated() =>
        Scenario()
            .WhenCreating(id => SampleItem.Create(id, " Primeiro ", Now))
            .Then(Created);

    [Fact]
    public void Criar_exige_nome() =>
        Scenario()
            .WhenCreating(id => SampleItem.Create(id, " ", Now))
            .ThenFails("O nome é obrigatório.");

    [Fact]
    public void Renomear_gera_SampleItemRenamed() =>
        Scenario()
            .Given(Created)
            .When(sampleItem => sampleItem.Rename("Segundo"))
            .Then(new SampleItemRenamed("Segundo"));

    [Fact]
    public void Renomear_para_o_mesmo_nome_nao_gera_evento() =>
        Scenario()
            .Given(Created)
            .When(sampleItem => sampleItem.Rename(" Primeiro "))
            .Then();

    [Fact]
    public void Renomear_exige_nome() =>
        Scenario()
            .Given(Created)
            .When(sampleItem => sampleItem.Rename(""))
            .ThenFails("O nome é obrigatório.");

    [Fact]
    public void Projecao_acompanha_os_eventos()
    {
        var id = Guid.NewGuid();
        var projection = new SampleItemSummaryProjection();

        var summary = projection.Evolve(projection.Evolve(null, id, Created), id, new SampleItemRenamed("Segundo"));

        Assert.Equal(new SampleItemSummary { Id = id, Name = "Segundo", CreatedAt = Now, Renames = 1 }, summary);
    }
}
