using SampleApp.Catalog;
using SampleApp.Ordering;
// romanctl:usings

namespace SampleApp.Api;

internal static class Modules
{
    public static IHostApplicationBuilder AddModules(this IHostApplicationBuilder builder)
    {
        builder.AddCatalogModule();
        builder.AddOrderingModule();
        // romanctl:modules
        return builder;
    }

    public static IEndpointRouteBuilder MapModules(this IEndpointRouteBuilder app)
    {
        app.MapCatalogEndpoints();
        app.MapOrderingEndpoints();
        // romanctl:endpoints
        return app;
    }
}
