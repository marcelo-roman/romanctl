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
    public void A_API_compoe_os_modulos_e_cada_um_tem_contratos()
    {
        var modules = Modules();

        Assert.True(modules.Count >= 2, "A API deveria registrar pelo menos dois módulos.");
        Assert.All(modules, module => Assert.NotNull(ContractsOf(module)));
    }

    [Fact]
    public void Um_modulo_so_conhece_os_contratos_dos_outros()
    {
        var modules = Modules();
        foreach (var module in modules)
        {
            var others = InternalTypesOf(modules.Where(other => other != module));

            var result = Types.InAssembly(module).ShouldNot().HaveDependencyOnAny(others).GetResult();

            Assert.True(result.IsSuccessful, $"{Name(module)} usa o interior de outro módulo: {Describe(result)}");
        }
    }

    [Fact]
    public void Contratos_nao_dependem_de_implementacao_nem_de_frameworks()
    {
        var modules = Modules();
        var forbidden = InternalTypesOf(modules).Concat(Frameworks).ToArray();
        foreach (var contracts in modules.Select(ContractsOf))
        {
            var result = Types.InAssembly(contracts).ShouldNot().HaveDependencyOnAny(forbidden).GetResult();

            Assert.True(result.IsSuccessful, $"{Name(contracts)} depende do que não deveria: {Describe(result)}");
        }
    }

    [Fact]
    public void Shared_nao_conhece_nenhum_modulo()
    {
        var modules = Modules().Select(Name).ToArray();

        var result = Types.InAssembly(Shared).ShouldNot().HaveDependencyOnAny(modules).GetResult();

        Assert.True(result.IsSuccessful, $"Shared depende de um módulo: {Describe(result)}");
    }

    [Fact]
    public void Dominio_de_cada_modulo_nao_depende_das_outras_camadas_nem_de_frameworks()
    {
        foreach (var module in Modules())
        {
            var name = Name(module);
            var layers = new[] { ".Application", ".Infrastructure", ".Endpoints" }.Select(layer => name + layer);

            var result = Types.InAssembly(module)
                .That().ResideInNamespace(name + ".Domain")
                .ShouldNot().HaveDependencyOnAny(layers.Concat(Frameworks).ToArray())
                .GetResult();

            Assert.True(result.IsSuccessful, $"O domínio de {name} depende de fora dele: {Describe(result)}");
        }
    }
}
