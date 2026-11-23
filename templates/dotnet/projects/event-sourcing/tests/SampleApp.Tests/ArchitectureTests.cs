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
    public void Domain_depends_on_no_other_layer() =>
        Assert.Empty(ProjectReferences(Domain));

    [Fact]
    public void Application_depends_only_on_Domain() =>
        Assert.Equal(["SampleApp.Domain"], ProjectReferences(Application));

    [Fact]
    public void Infrastructure_is_not_referenced_by_Domain_or_Application() =>
        Assert.DoesNotContain("SampleApp.Infrastructure", ProjectReferences(Domain).Concat(ProjectReferences(Application)));

    [Fact]
    public void Domain_uses_no_frameworks() =>
        Assert.DoesNotContain(References(Domain), name => name.StartsWith("Microsoft.Extensions") || name.StartsWith("Marten"));

    [Fact]
    public void Application_does_not_know_the_concrete_event_store() =>
        Assert.DoesNotContain(References(Application), name => name.StartsWith("Marten") || name.StartsWith("JasperFx") || name.StartsWith("Npgsql"));
}
