using SampleApp.Api.Data;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Features.SampleItems;

public static class ListSampleItems
{
    public const int DefaultLimit = 50;
    public const int MaxLimit = 200;

    public sealed record Response(Guid Id, string Name, DateTimeOffset CreatedAt);

    public static void MapEndpoint(IEndpointRouteBuilder app) =>
        app.MapGet("/sample-items", HandleAsync).WithTags("SampleItems");

    private static async Task<Ok<List<Response>>> HandleAsync(int? limit, AppDbContext db, CancellationToken cancellationToken)
    {
        var sampleItems = await db.SampleItems.AsNoTracking()
            .OrderByDescending(sampleItem => sampleItem.CreatedAt)
            .Take(Math.Clamp(limit ?? DefaultLimit, 1, MaxLimit))
            .Select(sampleItem => new Response(sampleItem.Id, sampleItem.Name, sampleItem.CreatedAt))
            .ToListAsync(cancellationToken);
        return TypedResults.Ok(sampleItems);
    }
}
