using SampleApp.Api.Data;
using SampleApp.Api.Domain.SampleItems;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Features.SampleItems;

public static class GetSampleItem
{
    public sealed record Response(Guid Id, string Name, DateTimeOffset CreatedAt)
    {
        public static Response From(SampleItem sampleItem) => new(sampleItem.Id, sampleItem.Name, sampleItem.CreatedAt);
    }

    public static void MapEndpoint(IEndpointRouteBuilder app) =>
        app.MapGet("/sample-items/{id:guid}", HandleAsync).WithTags("SampleItems");

    private static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(
        Guid id, AppDbContext db, CancellationToken cancellationToken)
    {
        var sampleItem = await db.SampleItems.AsNoTracking().FirstOrDefaultAsync(sampleItem => sampleItem.Id == id, cancellationToken);
        if (sampleItem is null)
            return TypedResults.Problem(statusCode: StatusCodes.Status404NotFound, title: "Não encontrado", detail: $"SampleItem {id} não encontrado.");

        return TypedResults.Ok(Response.From(sampleItem));
    }
}
