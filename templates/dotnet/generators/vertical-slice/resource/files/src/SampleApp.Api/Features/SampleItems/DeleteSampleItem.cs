using SampleApp.Api.Data;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Features.SampleItems;

public static class DeleteSampleItem
{
    public static void MapEndpoint(IEndpointRouteBuilder app) =>
        app.MapDelete("/sample-items/{id:guid}", HandleAsync).WithTags("SampleItems");

    private static async Task<Results<NoContent, ProblemHttpResult>> HandleAsync(
        Guid id, AppDbContext db, CancellationToken cancellationToken)
    {
        var sampleItem = await db.SampleItems.FirstOrDefaultAsync(sampleItem => sampleItem.Id == id, cancellationToken);
        if (sampleItem is null)
            return TypedResults.Problem(statusCode: StatusCodes.Status404NotFound, title: "Não encontrado", detail: $"SampleItem {id} não encontrado.");

        db.SampleItems.Remove(sampleItem);
        await db.SaveChangesAsync(cancellationToken);
        return TypedResults.NoContent();
    }
}
