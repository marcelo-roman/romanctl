using System.Reflection;
using {{.Name}}.Application;
using {{.Name}}.Domain.Orders;

namespace {{.Name}}.Tests;

/// <summary>Falha se alguma camada passar a depender de uma camada mais externa.</summary>
public class ArchitectureTests
{
    private static IEnumerable<string> ProjectReferences(Assembly assembly) =>
        assembly.GetReferencedAssemblies()
            .Select(reference => reference.Name!)
            .Where(name => name.StartsWith("{{.Name}}."));

    [Fact]
    public void Domain_nao_depende_de_nenhuma_camada() =>
        Assert.Empty(ProjectReferences(typeof(Order).Assembly));

    [Fact]
    public void Application_depende_apenas_de_Domain() =>
        Assert.Equal(["{{.Name}}.Domain"], ProjectReferences(typeof(DependencyInjection).Assembly));

    [Fact]
    public void Domain_nao_usa_frameworks()
    {
        var external = typeof(Order).Assembly.GetReferencedAssemblies()
            .Select(reference => reference.Name!)
            .Where(name => name.StartsWith("Microsoft.Extensions") || name.StartsWith("Microsoft.EntityFrameworkCore"));

        Assert.Empty(external);
    }
}
