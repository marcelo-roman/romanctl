using SampleApp.Api.Data;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.EntityFrameworkCore;

namespace SampleApp.Api.Features.SampleItems;

public static class RenameSampleItem
{
    public sealed record Request(string? Name);

    public static void MapEndpoint(IEndpointRouteBuilder app) =>
        app.MapPut("/sample-items/{id:guid}", HandleAsync).WithTags("SampleItems");

    private static Dictionary<string, string[]> Validate(Request request)
    {
        var errors = new Dictionary<string, string[]>();
        if (request.Name is null)
            errors["name"] = ["Informe o nome."];
        return errors;
    }

    private static async Task<Results<NoContent, ValidationProblem, ProblemHttpResult>> HandleAsync(
        Guid id, Request request, AppDbContext db, CancellationToken cancellationToken)
    {
        var errors = Validate(request);
        if (errors.Count > 0)
            return TypedResults.ValidationProblem(errors);

        var sampleItem = await db.SampleItems.FirstOrDefaultAsync(sampleItem => sampleItem.Id == id, cancellationToken);
        if (sampleItem is null)
            return TypedResults.Problem(statusCode: StatusCodes.Status404NotFound, title: "Não encontrado", detail: $"SampleItem {id} não encontrado.");

        sampleItem.Rename(request.Name!);
        await db.SaveChangesAsync(cancellationToken);
        return TypedResults.NoContent();
    }
}
