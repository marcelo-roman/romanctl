using System.Reflection;
using SampleApp.Application;
using SampleApp.Domain.Orders;

namespace SampleApp.Tests;

public class ArchitectureTests
{
    private static readonly Assembly Domain = typeof(Order).Assembly;
    private static readonly Assembly Application = typeof(DependencyInjection).Assembly;

    private static IEnumerable<string> References(Assembly assembly) =>
        assembly.GetReferencedAssemblies().Select(reference => reference.Name!);

    private static IEnumerable<string> ProjectReferences(Assembly assembly) =>
        References(assembly).Where(name => name.StartsWith("SampleApp."));

    [Fact]
    public void Domain_nao_depende_de_nenhuma_camada() =>
        Assert.Empty(ProjectReferences(Domain));

    [Fact]
    public void Application_depende_apenas_de_Domain() =>
        Assert.Equal(["SampleApp.Domain"], ProjectReferences(Application));

    [Fact]
    public void Infrastructure_nao_e_referenciada_por_Domain_nem_Application() =>
        Assert.DoesNotContain("SampleApp.Infrastructure", ProjectReferences(Domain).Concat(ProjectReferences(Application)));

    [Fact]
    public void Domain_nao_usa_frameworks() =>
        Assert.DoesNotContain(References(Domain), name => name.StartsWith("Microsoft.Extensions") || name.StartsWith("Marten"));

    [Fact]
    public void Application_nao_conhece_o_event_store_concreto() =>
        Assert.DoesNotContain(References(Application), name => name.StartsWith("Marten") || name.StartsWith("JasperFx") || name.StartsWith("Npgsql"));
}
