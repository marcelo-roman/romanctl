using SampleApp.Api.Data;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Features.Orders;

public static class PlaceOrder
{
    public static void MapEndpoint(IEndpointRouteBuilder app) =>
        app.MapPost("/orders/{id:guid}/place", HandleAsync).WithTags("Orders");

    private static async Task<Results<NoContent, ProblemHttpResult>> HandleAsync(
        Guid id, AppDbContext db, CancellationToken cancellationToken)
    {
        var order = await db.Orders.FirstOrDefaultAsync(order => order.Id == id, cancellationToken);
        if (order is null)
            return TypedResults.Problem(statusCode: StatusCodes.Status404NotFound, title: "Não encontrado", detail: $"Pedido {id} não encontrado.");

        order.Place();
        await db.SaveChangesAsync(cancellationToken);
        return TypedResults.NoContent();
    }
}
