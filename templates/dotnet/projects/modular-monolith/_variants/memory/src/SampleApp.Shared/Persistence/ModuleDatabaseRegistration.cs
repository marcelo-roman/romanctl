using SampleApp.Shared.Persistence.InMemory;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

namespace SampleApp.Shared.Persistence;

public static class ModuleDatabaseRegistration
{
    public static IHostApplicationBuilder AddModuleDatabase(this IHostApplicationBuilder builder)
    {
        builder.Services.AddSingleton<IModuleDatabase>(new InMemoryModuleDatabase("sample-app"));
        return builder;
    }
}
