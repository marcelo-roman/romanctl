using SampleApp.Application.Messaging;
using SampleApp.Application.Orders.Commands;
using SampleApp.Application.Orders.Queries;

namespace SampleApp.Api;

internal static class OrderEndpoints
{
    public sealed record CreateOrderRequest(string CustomerName);

    public sealed record AddItemRequest(string ProductName, int Quantity, decimal UnitPrice, string? Currency);

    public static void MapOrderEndpoints(this IEndpointRouteBuilder app)
    {
        var orders = app.MapGroup("/orders");

        orders.MapPost("/", async (CreateOrderRequest request, IDispatcher dispatcher, CancellationToken ct) =>
        {
            var id = await dispatcher.SendAsync(new CreateOrder(request.CustomerName), ct);
            return Results.Created($"/orders/{id}", new { id });
        });

        orders.MapPost("/{id:guid}/items", async (Guid id, AddItemRequest request, IDispatcher dispatcher, CancellationToken ct) =>
        {
            await dispatcher.SendAsync(new AddOrderItem(id, request.ProductName, request.Quantity, request.UnitPrice, request.Currency), ct);
            return Results.NoContent();
        });

        orders.MapPost("/{id:guid}/place", async (Guid id, IDispatcher dispatcher, CancellationToken ct) =>
        {
            await dispatcher.SendAsync(new PlaceOrder(id), ct);
            return Results.NoContent();
        });

        orders.MapGet("/{id:guid}", async (Guid id, IDispatcher dispatcher, CancellationToken ct) =>
            Results.Ok(await dispatcher.QueryAsync(new GetOrderById(id), ct)));

        orders.MapGet("/", async (int? limit, IDispatcher dispatcher, CancellationToken ct) =>
            Results.Ok(await dispatcher.QueryAsync(new ListOrders(limit ?? 50), ct)));
    }
}
