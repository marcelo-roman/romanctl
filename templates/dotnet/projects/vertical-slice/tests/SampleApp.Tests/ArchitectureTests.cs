using System.Reflection;
using SampleApp.Api.Data;
using Microsoft.AspNetCore.Routing;
using Mono.Cecil;
using NetArchTest.Rules;

namespace SampleApp.Tests;

public class ArchitectureTests
{
    private const string FeaturesNamespace = "SampleApp.Api.Features";

    private static readonly Assembly Api = typeof(AppDbContext).Assembly;

    private sealed class PartOfSlice(Type slice) : ICustomRule
    {
        public bool MeetsRule(TypeDefinition type) =>
            type.FullName == slice.FullName || type.FullName.StartsWith(slice.FullName + "/");
    }

    private static IReadOnlyList<Type> Slices() =>
        Api.GetTypes()
            .Where(type => type.Namespace?.StartsWith(FeaturesNamespace + ".") == true)
            .Where(type => !type.IsNested && type.IsAbstract && type.IsSealed)
            .ToList();

    private static string Describe(TestResult result) =>
        string.Join(", ", result.FailingTypeNames ?? []);

    [Fact]
    public void Every_slice_maps_its_own_endpoint()
    {
        Assert.NotEmpty(Slices());
        Assert.All(Slices(), slice =>
            Assert.NotNull(slice.GetMethod("MapEndpoint", BindingFlags.Public | BindingFlags.Static, [typeof(IEndpointRouteBuilder)])));
    }

    [Fact]
    public void Slices_do_not_depend_on_each_other()
    {
        var slices = Slices();
        foreach (var slice in slices)
        {
            var others = slices.Where(other => other != slice).Select(other => other.FullName!).ToArray();

            var result = Types.InAssembly(Api)
                .That().MeetCustomRule(new PartOfSlice(slice))
                .ShouldNot().HaveDependencyOnAny(others)
                .GetResult();

            Assert.True(result.IsSuccessful, $"{slice.Name} depends on another slice: {Describe(result)}");
        }
    }

    [Fact]
    public void Domain_depends_on_no_slice_or_framework()
    {
        var result = Types.InAssembly(Api)
            .That().ResideInNamespace("SampleApp.Api.Domain")
            .ShouldNot().HaveDependencyOnAny(FeaturesNamespace, "SampleApp.Api.Data", "Microsoft.AspNetCore", "Microsoft.EntityFrameworkCore")
            .GetResult();

        Assert.True(result.IsSuccessful, $"The domain depends on something outside it: {Describe(result)}");
    }
}
