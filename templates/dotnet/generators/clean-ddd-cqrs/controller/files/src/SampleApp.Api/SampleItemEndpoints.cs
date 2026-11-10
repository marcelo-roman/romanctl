using SampleApp.Application.Messaging;
using SampleApp.Application.SampleItems.Commands;
using SampleApp.Application.SampleItems.Queries;

namespace SampleApp.Api;

internal static class SampleItemEndpoints
{
    public sealed record SampleItemRequest(string Name);

    public static void MapSampleItemEndpoints(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/sample-items");

        group.MapPost("/", async (SampleItemRequest request, IDispatcher dispatcher, CancellationToken ct) =>
        {
            var id = await dispatcher.SendAsync(new CreateSampleItem(request.Name), ct);
            return Results.Created($"/sample-items/{id}", new { id });
        });

        group.MapPut("/{id:guid}", async (Guid id, SampleItemRequest request, IDispatcher dispatcher, CancellationToken ct) =>
        {
            await dispatcher.SendAsync(new RenameSampleItem(id, request.Name), ct);
            return Results.NoContent();
        });

        group.MapDelete("/{id:guid}", async (Guid id, IDispatcher dispatcher, CancellationToken ct) =>
        {
            await dispatcher.SendAsync(new DeleteSampleItem(id), ct);
            return Results.NoContent();
        });

        group.MapGet("/{id:guid}", async (Guid id, IDispatcher dispatcher, CancellationToken ct) =>
            Results.Ok(await dispatcher.QueryAsync(new GetSampleItemById(id), ct)));

        group.MapGet("/", async (int? limit, IDispatcher dispatcher, CancellationToken ct) =>
            Results.Ok(await dispatcher.QueryAsync(new ListSampleItems(limit ?? 50), ct)));
    }
}
