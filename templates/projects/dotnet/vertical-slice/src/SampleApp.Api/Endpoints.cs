namespace SampleApp.Api;

internal static class Endpoints
{
    public static void MapFeatures(this IEndpointRouteBuilder app)
    {
        app.MapGet("/health", () => Results.Ok(new { status = "ok" }));
        Features.Orders.CreateOrder.MapEndpoint(app);
        Features.Orders.AddItem.MapEndpoint(app);
        Features.Orders.PlaceOrder.MapEndpoint(app);
        Features.Orders.GetOrder.MapEndpoint(app);
        Features.Orders.ListOrders.MapEndpoint(app);
        // romanctl:endpoints
    }
}
