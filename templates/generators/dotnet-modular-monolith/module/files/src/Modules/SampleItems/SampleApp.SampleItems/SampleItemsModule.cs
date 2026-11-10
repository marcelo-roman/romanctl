using SampleApp.SampleItems.Application;
using SampleApp.SampleItems.Contracts;
using SampleApp.SampleItems.Domain;
using SampleApp.SampleItems.Endpoints;
using SampleApp.SampleItems.Infrastructure;
using SampleApp.Shared.Persistence;
using Microsoft.AspNetCore.Routing;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

namespace SampleApp.SampleItems;

public static class SampleItemsModule
{
    public static IHostApplicationBuilder AddSampleItemsModule(this IHostApplicationBuilder builder)
    {
        var services = builder.Services;
        services.AddModuleDbContext<SampleItemsDbContext>(SampleItemsDbContext.Schema);
        services.AddScoped<ISampleItemRepository, SampleItemRepository>();
        services.AddScoped<CreateSampleItemHandler>();
        services.AddScoped<RenameSampleItemHandler>();
        services.AddScoped<GetSampleItemHandler>();
        services.AddScoped<ListSampleItemsHandler>();
        services.AddScoped<DeleteSampleItemHandler>();
        services.AddScoped<ISampleItemsApi, SampleItemsApi>();
        // romanctl:services
        return builder;
    }

    public static IEndpointRouteBuilder MapSampleItemsEndpoints(this IEndpointRouteBuilder app)
    {
        app.MapSampleItemEndpoints();
        // romanctl:endpoints
        return app;
    }
}
