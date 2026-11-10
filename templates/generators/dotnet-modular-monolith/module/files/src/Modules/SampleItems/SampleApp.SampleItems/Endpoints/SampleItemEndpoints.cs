using SampleApp.SampleItems.Application;
using SampleApp.Shared.Web;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Routing;

namespace SampleApp.SampleItems.Endpoints;

internal static class SampleItemEndpoints
{
    public const string Route = "/sample-items";

    public sealed record CreateSampleItemRequest(string Name);

    public sealed record RenameSampleItemRequest(string Name);

    public static void MapSampleItemEndpoints(this IEndpointRouteBuilder app)
    {
        var sampleItems = app.MapGroup(Route).WithTags("SampleItems");

        sampleItems.MapPost("/", async (CreateSampleItemRequest request, CreateSampleItemHandler handler, CancellationToken ct) =>
        {
            var id = await handler.HandleAsync(new CreateSampleItem(request.Name), ct);
            return Results.Created($"{Route}/{id}", new CreatedResponse(id));
        });

        sampleItems.MapPut("/{id:guid}", async (Guid id, RenameSampleItemRequest request, RenameSampleItemHandler handler, CancellationToken ct) =>
        {
            await handler.HandleAsync(new RenameSampleItem(id, request.Name), ct);
            return Results.NoContent();
        });

        sampleItems.MapGet("/{id:guid}", async (Guid id, GetSampleItemHandler handler, CancellationToken ct) =>
            Results.Ok(await handler.HandleAsync(id, ct)));

        sampleItems.MapGet("/", async (int? limit, ListSampleItemsHandler handler, CancellationToken ct) =>
            Results.Ok(await handler.HandleAsync(limit, ct)));

        sampleItems.MapDelete("/{id:guid}", async (Guid id, DeleteSampleItemHandler handler, CancellationToken ct) =>
        {
            await handler.HandleAsync(id, ct);
            return Results.NoContent();
        });
    }
}
