using SampleApp.Api.Data;
using SampleApp.Api.Domain.SampleItems;
using Microsoft.AspNetCore.Http.HttpResults;

namespace SampleApp.Api.Features.SampleItems;

public static class CreateSampleItem
{
    public sealed record Request(string? Name);

    public sealed record Response(Guid Id);

    public static void MapEndpoint(IEndpointRouteBuilder app) =>
        app.MapPost("/sample-items", HandleAsync).WithTags("SampleItems");

    private static Dictionary<string, string[]> Validate(Request request)
    {
        var errors = new Dictionary<string, string[]>();
        if (request.Name is null)
            errors["name"] = ["Informe o nome."];
        return errors;
    }

    private static async Task<Results<Created<Response>, ValidationProblem>> HandleAsync(
        Request request, AppDbContext db, TimeProvider clock, CancellationToken cancellationToken)
    {
        var errors = Validate(request);
        if (errors.Count > 0)
            return TypedResults.ValidationProblem(errors);

        var sampleItem = SampleItem.Create(request.Name!, clock.GetUtcNow());
        db.SampleItems.Add(sampleItem);
        await db.SaveChangesAsync(cancellationToken);
        return TypedResults.Created($"/sample-items/{sampleItem.Id}", new Response(sampleItem.Id));
    }
}
