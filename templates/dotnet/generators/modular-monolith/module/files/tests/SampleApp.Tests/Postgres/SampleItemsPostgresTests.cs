using System.Net.Http.Json;
using SampleApp.SampleItems.Application;
using SampleApp.SampleItems.Endpoints;
using SampleApp.SampleItems.Infrastructure;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Tests.Postgres;

[Collection(PostgresCollection.Name)]
public class SampleItemsPostgresTests
{
    [PostgresFact]
    public async Task SampleItem_round_trips_through_the_module_schema()
    {
        await using var factory = new WebApplicationFactory<Program>();
        using var client = factory.CreateClient();

        var id = await TestApi.CreateAsync(client, SampleItemEndpoints.Route, new { name = "Example" });
        var found = await client.GetFromJsonAsync<SampleItemResponse>($"{SampleItemEndpoints.Route}/{id}");

        Assert.Equal("Example", found!.Name);
        await using var scope = factory.Services.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<SampleItemsDbContext>();
        var tables = await db.Database
            .SqlQuery<string>($"select table_name as \"Value\" from information_schema.tables where table_schema = {SampleItemsDbContext.Schema}")
            .ToListAsync();
        Assert.Contains("sample_items", tables);
    }
}
