using SampleApp.Api.Data;
using SampleApp.Api.Domain.Orders;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Features.Orders;

public static class GetOrder
{
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
        app.MapGet("/orders/{id:guid}", HandleAsync).WithTags("Orders");

    private static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(
        Guid id, AppDbContext db, CancellationToken cancellationToken)
    {
        var order = await db.Orders.AsNoTracking().FirstOrDefaultAsync(order => order.Id == id, cancellationToken);
        if (order is null)
            return TypedResults.Problem(statusCode: StatusCodes.Status404NotFound, title: "Não encontrado", detail: $"Pedido {id} não encontrado.");

        return TypedResults.Ok(Response.From(order));
    }
}
