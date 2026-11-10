using SampleApp.Api.Data;
using SampleApp.Api.Domain.Orders;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Features.Orders;

public static class ListOrders
{
    public const int DefaultLimit = 50;
    public const int MaxLimit = 200;

    public sealed record ItemResponse(string ProductName, int Quantity, decimal UnitPrice, decimal Subtotal);

    public sealed record Response(
        Guid Id,
        string CustomerName,
        string Status,
        decimal Total,
        string Currency,
        DateTimeOffset CreatedAt,
        IReadOnlyList<ItemResponse> Items)
    {
        public static Response From(Order order)
        {
            var total = order.Total;
            return new Response(
                order.Id,
                order.CustomerName,
                order.Status.ToString(),
                total.Amount,
                total.Currency,
                order.CreatedAt,
                order.Items
                    .Select(item => new ItemResponse(item.ProductName, item.Quantity, item.UnitPrice.Amount, item.Subtotal.Amount))
                    .ToList());
        }
    }

    public static void MapEndpoint(IEndpointRouteBuilder app) =>
        app.MapGet("/orders", HandleAsync).WithTags("Orders");

    private static async Task<Ok<List<Response>>> HandleAsync(int? limit, AppDbContext db, CancellationToken cancellationToken)
    {
        var orders = await db.Orders.AsNoTracking()
            .OrderByDescending(order => order.CreatedAt)
            .Take(Math.Clamp(limit ?? DefaultLimit, 1, MaxLimit))
            .ToListAsync(cancellationToken);
        return TypedResults.Ok(orders.Select(Response.From).ToList());
    }
}
