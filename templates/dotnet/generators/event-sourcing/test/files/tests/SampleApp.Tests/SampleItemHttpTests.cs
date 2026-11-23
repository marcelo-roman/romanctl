using System.Net;
using System.Net.Http.Json;
using SampleApp.Application.SampleItems;

namespace SampleApp.Tests;

public class SampleItemHttpTests(ApiFactory factory) : IClassFixture<ApiFactory>
{
    private sealed record Created(Guid Id);

    private sealed record StreamEvent(long Version, string Type);

    private readonly HttpClient _client = factory.CreateClient();

    [Fact]
    public async Task Creates_renames_and_reads_through_the_projection()
    {
        var created = await _client.PostAsJsonAsync("/sample-items", new { name = "First" });
        Assert.Equal(HttpStatusCode.Created, created.StatusCode);
        var id = (await created.Content.ReadFromJsonAsync<Created>())!.Id;

        var renamed = await _client.PutAsJsonAsync($"/sample-items/{id}", new { name = "Second" });
        Assert.Equal(HttpStatusCode.NoContent, renamed.StatusCode);

        var summary = await _client.GetFromJsonAsync<SampleItemSummary>($"/sample-items/{id}");
        Assert.Equal("Second", summary!.Name);
        Assert.Contains(await _client.GetFromJsonAsync<List<SampleItemSummary>>("/sample-items?limit=200") ?? [], item => item.Id == id);

        var events = (await _client.GetFromJsonAsync<List<StreamEvent>>($"/sample-items/{id}/events"))!;
        Assert.Equal(["SampleItemCreated", "SampleItemRenamed"], events.Select(e => e.Type));
    }

    [Fact]
    public async Task Empty_name_becomes_422_and_unknown_becomes_404()
    {
        Assert.Equal(HttpStatusCode.UnprocessableEntity, (await _client.PostAsJsonAsync("/sample-items", new { name = " " })).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await _client.GetAsync($"/sample-items/{Guid.NewGuid()}")).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await _client.PutAsJsonAsync($"/sample-items/{Guid.NewGuid()}", new { name = "X" })).StatusCode);
    }
}
