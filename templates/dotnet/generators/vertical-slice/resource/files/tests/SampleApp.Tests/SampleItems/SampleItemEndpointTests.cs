using System.Net;
using System.Net.Http.Json;
using SampleApp.Api.Features.SampleItems;

namespace SampleApp.Tests.SampleItems;

public class SampleItemEndpointTests
{
    private static async Task<Guid> CreateAsync(HttpClient client, string name)
    {
        var response = await client.PostAsJsonAsync("/sample-items", new { name });
        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        return (await response.Content.ReadFromJsonAsync<CreateSampleItem.Response>())!.Id;
    }

    [Fact]
    public async Task Creates_renames_gets_lists_and_deletes()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        var id = await CreateAsync(client, "Example");
        Assert.Equal(HttpStatusCode.NoContent, (await client.PutAsJsonAsync($"/sample-items/{id}", new { name = "Other" })).StatusCode);

        var found = await client.GetFromJsonAsync<GetSampleItem.Response>($"/sample-items/{id}");
        Assert.Equal("Other", found!.Name);
        var listed = await client.GetFromJsonAsync<List<ListSampleItems.Response>>("/sample-items?limit=10");
        Assert.Equal(id, Assert.Single(listed!).Id);

        Assert.Equal(HttpStatusCode.NoContent, (await client.DeleteAsync($"/sample-items/{id}")).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await client.GetAsync($"/sample-items/{id}")).StatusCode);
    }

    [Fact]
    public async Task Answers_400_422_and_404()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var id = await CreateAsync(client, "Example");

        Assert.Equal(HttpStatusCode.BadRequest, (await client.PostAsJsonAsync("/sample-items", new { })).StatusCode);
        Assert.Equal(HttpStatusCode.UnprocessableEntity, (await client.PostAsJsonAsync("/sample-items", new { name = " " })).StatusCode);
        Assert.Equal(HttpStatusCode.UnprocessableEntity, (await client.PutAsJsonAsync($"/sample-items/{id}", new { name = "" })).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await client.PutAsJsonAsync($"/sample-items/{Guid.NewGuid()}", new { name = "Other" })).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await client.DeleteAsync($"/sample-items/{Guid.NewGuid()}")).StatusCode);
    }
}
