using SampleApp.Catalog.Application;
using SampleApp.Catalog.Contracts;
using SampleApp.Catalog.Domain;
using SampleApp.Catalog.Endpoints;
using SampleApp.Catalog.Infrastructure;
using SampleApp.Ordering.Contracts;
using SampleApp.Shared.Messaging;
using SampleApp.Shared.Persistence;
using Microsoft.AspNetCore.Routing;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

namespace SampleApp.Catalog;

public static class CatalogModule
{
    public static IHostApplicationBuilder AddCatalogModule(this IHostApplicationBuilder builder)
    {
        var services = builder.Services;
        services.AddModuleDbContext<CatalogDbContext>(CatalogDbContext.Schema);
        services.AddScoped<IProductRepository, ProductRepository>();
        services.AddScoped<CreateProductHandler>();
        services.AddScoped<GetProductHandler>();
        services.AddScoped<ListProductsHandler>();
        services.AddScoped<ICatalogApi, CatalogApi>();
        services.AddIntegrationEventHandler<OrderPlaced, RecordSalesOnOrderPlaced>();
        // romanctl:services
        return builder;
    }

    public static IEndpointRouteBuilder MapCatalogEndpoints(this IEndpointRouteBuilder app)
    {
        app.MapProductEndpoints();
        // romanctl:endpoints
        return app;
    }
}
