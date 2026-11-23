using System.Reflection;
using NetArchTest.Rules;

namespace SampleApp.Tests.Architecture;

public class ModuleBoundaryTests
{
    private const string Root = "SampleApp";
    private const string ContractsSuffix = ".Contracts";
    private const string SharedName = Root + ".Shared";

    private static readonly string[] Frameworks = ["Microsoft.AspNetCore", "Microsoft.EntityFrameworkCore"];

    private static readonly Assembly Shared = Assembly.Load(SharedName);

    private static IReadOnlyList<Assembly> Modules() =>
        typeof(Program).Assembly.GetReferencedAssemblies()
            .Select(reference => reference.Name!)
            .Where(name => name.StartsWith(Root + ".") && name != SharedName && !name.EndsWith(ContractsSuffix))
            .Order()
            .Select(Assembly.Load)
            .ToList();

    private static Assembly ContractsOf(Assembly module) => Assembly.Load(Name(module) + ContractsSuffix);

    private static string Name(Assembly assembly) => assembly.GetName().Name!;

    private static string[] InternalTypesOf(IEnumerable<Assembly> modules) =>
        modules
            .SelectMany(module => module.GetTypes().Where(type => !type.IsNested && type.Namespace?.StartsWith(Name(module)) == true))
            .Select(type => type.FullName!)
            .ToArray();

    private static string Describe(TestResult result) =>
        string.Join(", ", result.FailingTypeNames ?? []);

    [Fact]
    public void The_API_composes_the_modules_and_each_has_contracts()
    {
        var modules = Modules();

        Assert.True(modules.Count >= 2, "The API should register at least two modules.");
        Assert.All(modules, module => Assert.NotNull(ContractsOf(module)));
    }

    [Fact]
    public void A_module_only_knows_the_contracts_of_the_others()
    {
        var modules = Modules();
        foreach (var module in modules)
        {
            var others = InternalTypesOf(modules.Where(other => other != module));

            var result = Types.InAssembly(module).ShouldNot().HaveDependencyOnAny(others).GetResult();

            Assert.True(result.IsSuccessful, $"{Name(module)} uses the internals of another module: {Describe(result)}");
        }
    }

    [Fact]
    public void Contracts_depend_on_no_implementation_or_framework()
    {
        var modules = Modules();
        var forbidden = InternalTypesOf(modules).Concat(Frameworks).ToArray();
        foreach (var contracts in modules.Select(ContractsOf))
        {
            var result = Types.InAssembly(contracts).ShouldNot().HaveDependencyOnAny(forbidden).GetResult();

            Assert.True(result.IsSuccessful, $"{Name(contracts)} depends on something it should not: {Describe(result)}");
        }
    }

    [Fact]
    public void Shared_knows_no_module()
    {
        var modules = Modules().Select(Name).ToArray();

        var result = Types.InAssembly(Shared).ShouldNot().HaveDependencyOnAny(modules).GetResult();

        Assert.True(result.IsSuccessful, $"Shared depends on a module: {Describe(result)}");
    }

    [Fact]
    public void Each_module_domain_depends_on_no_other_layer_or_framework()
    {
        foreach (var module in Modules())
        {
            var name = Name(module);
            var layers = new[] { ".Application", ".Infrastructure", ".Endpoints" }.Select(layer => name + layer);

            var result = Types.InAssembly(module)
                .That().ResideInNamespace(name + ".Domain")
                .ShouldNot().HaveDependencyOnAny(layers.Concat(Frameworks).ToArray())
                .GetResult();

            Assert.True(result.IsSuccessful, $"The domain of {name} depends on something outside it: {Describe(result)}");
        }
    }
}
