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
    public async Task Cria_renomeia_consulta_lista_e_remove()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        var id = await CreateAsync(client, "Exemplo");
        Assert.Equal(HttpStatusCode.NoContent, (await client.PutAsJsonAsync($"/sample-items/{id}", new { name = "Outro" })).StatusCode);

        var found = await client.GetFromJsonAsync<GetSampleItem.Response>($"/sample-items/{id}");
        Assert.Equal("Outro", found!.Name);
        var listed = await client.GetFromJsonAsync<List<ListSampleItems.Response>>("/sample-items?limit=10");
        Assert.Equal(id, Assert.Single(listed!).Id);

        Assert.Equal(HttpStatusCode.NoContent, (await client.DeleteAsync($"/sample-items/{id}")).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await client.GetAsync($"/sample-items/{id}")).StatusCode);
    }

    [Fact]
    public async Task Responde_400_422_e_404()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var id = await CreateAsync(client, "Exemplo");

        Assert.Equal(HttpStatusCode.BadRequest, (await client.PostAsJsonAsync("/sample-items", new { })).StatusCode);
        Assert.Equal(HttpStatusCode.UnprocessableEntity, (await client.PostAsJsonAsync("/sample-items", new { name = " " })).StatusCode);
        Assert.Equal(HttpStatusCode.UnprocessableEntity, (await client.PutAsJsonAsync($"/sample-items/{id}", new { name = "" })).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await client.PutAsJsonAsync($"/sample-items/{Guid.NewGuid()}", new { name = "Outro" })).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await client.DeleteAsync($"/sample-items/{Guid.NewGuid()}")).StatusCode);
    }
}
