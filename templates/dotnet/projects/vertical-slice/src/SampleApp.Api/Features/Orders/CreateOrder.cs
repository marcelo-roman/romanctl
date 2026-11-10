using SampleApp.Api.Data;
using SampleApp.Api.Domain.Orders;
using Microsoft.AspNetCore.Http.HttpResults;

namespace SampleApp.Api.Features.Orders;

public static class CreateOrder
{
    public sealed record Request(string? CustomerName);

    public sealed record Response(Guid Id);

    public static void MapEndpoint(IEndpointRouteBuilder app) =>
        app.MapPost("/orders", HandleAsync).WithTags("Orders");

    private static Dictionary<string, string[]> Validate(Request request)
    {
        var errors = new Dictionary<string, string[]>();
        if (request.CustomerName is null)
            errors["customerName"] = ["Informe o nome do cliente."];
        return errors;
    }

    private static async Task<Results<Created<Response>, ValidationProblem>> HandleAsync(
        Request request, AppDbContext db, TimeProvider clock, CancellationToken cancellationToken)
    {
        var errors = Validate(request);
        if (errors.Count > 0)
            return TypedResults.ValidationProblem(errors);

        var order = Order.Create(request.CustomerName!, clock.GetUtcNow());
        db.Orders.Add(order);
        await db.SaveChangesAsync(cancellationToken);
        return TypedResults.Created($"/orders/{order.Id}", new Response(order.Id));
    }
}
