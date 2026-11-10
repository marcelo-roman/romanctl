using SampleApp.Api.Data;
using SampleApp.Api.Domain.Orders;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Features.Orders;

public static class AddItem
{
    public sealed record Request(string? ProductName, int? Quantity, decimal? UnitPrice, string? Currency);

    public static void MapEndpoint(IEndpointRouteBuilder app) =>
        app.MapPost("/orders/{id:guid}/items", HandleAsync).WithTags("Orders");

    private static Dictionary<string, string[]> Validate(Request request)
    {
        var errors = new Dictionary<string, string[]>();
        if (request.ProductName is null)
            errors["productName"] = ["Informe o nome do produto."];
        if (request.Quantity is null)
            errors["quantity"] = ["Informe a quantidade."];
        if (request.UnitPrice is null)
            errors["unitPrice"] = ["Informe o preço unitário."];
        return errors;
    }

    private static async Task<Results<NoContent, ValidationProblem, ProblemHttpResult>> HandleAsync(
        Guid id, Request request, AppDbContext db, CancellationToken cancellationToken)
    {
        var errors = Validate(request);
        if (errors.Count > 0)
            return TypedResults.ValidationProblem(errors);

        var order = await db.Orders.FirstOrDefaultAsync(order => order.Id == id, cancellationToken);
        if (order is null)
            return TypedResults.Problem(statusCode: StatusCodes.Status404NotFound, title: "Não encontrado", detail: $"Pedido {id} não encontrado.");

        var unitPrice = new Money(request.UnitPrice!.Value, request.Currency ?? Order.DefaultCurrency);
        order.AddItem(request.ProductName!, request.Quantity!.Value, unitPrice);
        await db.SaveChangesAsync(cancellationToken);
        return TypedResults.NoContent();
    }
}
