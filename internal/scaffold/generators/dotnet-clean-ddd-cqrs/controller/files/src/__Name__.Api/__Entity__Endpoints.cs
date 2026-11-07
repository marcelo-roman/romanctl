using {{.Name}}.Application.Messaging;
using {{.Name}}.Application.{{.EntityPlural}}.Commands;
using {{.Name}}.Application.{{.EntityPlural}}.Queries;

namespace {{.Name}}.Api;

internal static class {{.Entity}}Endpoints
{
    public sealed record {{.Entity}}Request(string Name);

    public static void Map{{.Entity}}Endpoints(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/{{.EntityPluralKebab}}");

        group.MapPost("/", async ({{.Entity}}Request request, IDispatcher dispatcher, CancellationToken ct) =>
        {
            var id = await dispatcher.SendAsync(new Create{{.Entity}}(request.Name), ct);
            return Results.Created($"/{{.EntityPluralKebab}}/{id}", new { id });
        });

        group.MapPut("/{id:guid}", async (Guid id, {{.Entity}}Request request, IDispatcher dispatcher, CancellationToken ct) =>
        {
            await dispatcher.SendAsync(new Rename{{.Entity}}(id, request.Name), ct);
            return Results.NoContent();
        });

        group.MapDelete("/{id:guid}", async (Guid id, IDispatcher dispatcher, CancellationToken ct) =>
        {
            await dispatcher.SendAsync(new Delete{{.Entity}}(id), ct);
            return Results.NoContent();
        });

        group.MapGet("/{id:guid}", async (Guid id, IDispatcher dispatcher, CancellationToken ct) =>
            Results.Ok(await dispatcher.QueryAsync(new Get{{.Entity}}ById(id), ct)));

        group.MapGet("/", async (int? limit, IDispatcher dispatcher, CancellationToken ct) =>
            Results.Ok(await dispatcher.QueryAsync(new List{{.EntityPlural}}(limit ?? 50), ct)));
    }
}
