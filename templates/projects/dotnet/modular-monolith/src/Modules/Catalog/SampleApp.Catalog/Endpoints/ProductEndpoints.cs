using SampleApp.Catalog.Application;
using SampleApp.Shared.Web;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Routing;

namespace SampleApp.Catalog.Endpoints;

internal static class ProductEndpoints
{
    public sealed record CreateProductRequest(string Name, decimal Price, string? Currency = null);

    public static void MapProductEndpoints(this IEndpointRouteBuilder app)
    {
        var products = app.MapGroup("/catalog/products").WithTags("Catalog");

        products.MapPost("/", async (CreateProductRequest request, CreateProductHandler handler, CancellationToken ct) =>
        {
            var id = await handler.HandleAsync(new CreateProduct(request.Name, request.Price, request.Currency), ct);
            return Results.Created($"/catalog/products/{id}", new CreatedResponse(id));
        });

        products.MapGet("/{id:guid}", async (Guid id, GetProductHandler handler, CancellationToken ct) =>
            Results.Ok(await handler.HandleAsync(id, ct)));

        products.MapGet("/", async (int? limit, ListProductsHandler handler, CancellationToken ct) =>
            Results.Ok(await handler.HandleAsync(limit, ct)));
    }
}
