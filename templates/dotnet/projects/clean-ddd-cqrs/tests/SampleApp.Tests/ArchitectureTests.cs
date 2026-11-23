using System.Reflection;
using SampleApp.Application;
using SampleApp.Domain.Orders;

namespace SampleApp.Tests;

/// <summary>Fails if any layer starts depending on an outer layer.</summary>
public class ArchitectureTests
{
    private static IEnumerable<string> ProjectReferences(Assembly assembly) =>
        assembly.GetReferencedAssemblies()
            .Select(reference => reference.Name!)
            .Where(name => name.StartsWith("SampleApp."));

    [Fact]
    public void Domain_depends_on_no_other_layer() =>
        Assert.Empty(ProjectReferences(typeof(Order).Assembly));

    [Fact]
    public void Application_depends_only_on_Domain() =>
        Assert.Equal(["SampleApp.Domain"], ProjectReferences(typeof(DependencyInjection).Assembly));

    [Fact]
    public void Domain_uses_no_frameworks()
    {
        var external = typeof(Order).Assembly.GetReferencedAssemblies()
            .Select(reference => reference.Name!)
            .Where(name => name.StartsWith("Microsoft.Extensions") || name.StartsWith("Microsoft.EntityFrameworkCore"));

        Assert.Empty(external);
    }
}
