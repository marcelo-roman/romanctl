using SampleApp.Ordering.Application;
using SampleApp.Ordering.Domain;
using SampleApp.Ordering.Endpoints;
using SampleApp.Ordering.Infrastructure;
using SampleApp.Shared.Persistence;
using Microsoft.AspNetCore.Routing;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

namespace SampleApp.Ordering;

public static class OrderingModule
{
    public static IHostApplicationBuilder AddOrderingModule(this IHostApplicationBuilder builder)
    {
        var services = builder.Services;
        services.AddModuleDbContext<OrderingDbContext>(OrderingDbContext.Schema);
        services.AddScoped<IOrderRepository, OrderRepository>();
        services.AddScoped<CreateOrderHandler>();
        services.AddScoped<AddOrderItemHandler>();
        services.AddScoped<PlaceOrderHandler>();
        services.AddScoped<GetOrderHandler>();
        services.AddScoped<ListOrdersHandler>();
        // romanctl:services
        return builder;
    }

    public static IEndpointRouteBuilder MapOrderingEndpoints(this IEndpointRouteBuilder app)
    {
        app.MapOrderEndpoints();
        // romanctl:endpoints
        return app;
    }
}
