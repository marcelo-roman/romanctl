using SampleApp.Ordering.Application;
using SampleApp.Shared.Web;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Routing;

namespace SampleApp.Ordering.Endpoints;

internal static class OrderEndpoints
{
    public sealed record CreateOrderRequest(string CustomerName);

    public sealed record AddItemRequest(Guid ProductId, int Quantity);

    public static void MapOrderEndpoints(this IEndpointRouteBuilder app)
    {
        var orders = app.MapGroup("/orders").WithTags("Ordering");

        orders.MapPost("/", async (CreateOrderRequest request, CreateOrderHandler handler, CancellationToken ct) =>
        {
            var id = await handler.HandleAsync(new CreateOrder(request.CustomerName), ct);
            return Results.Created($"/orders/{id}", new CreatedResponse(id));
        });

        orders.MapPost("/{id:guid}/items", async (Guid id, AddItemRequest request, AddOrderItemHandler handler, CancellationToken ct) =>
        {
            await handler.HandleAsync(new AddOrderItem(id, request.ProductId, request.Quantity), ct);
            return Results.NoContent();
        });

        orders.MapPost("/{id:guid}/place", async (Guid id, PlaceOrderHandler handler, CancellationToken ct) =>
        {
            await handler.HandleAsync(new PlaceOrder(id), ct);
            return Results.NoContent();
        });

        orders.MapGet("/{id:guid}", async (Guid id, GetOrderHandler handler, CancellationToken ct) =>
            Results.Ok(await handler.HandleAsync(id, ct)));

        orders.MapGet("/", async (int? limit, ListOrdersHandler handler, CancellationToken ct) =>
            Results.Ok(await handler.HandleAsync(limit, ct)));
    }
}
