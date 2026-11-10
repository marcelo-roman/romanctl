using System.Net;
using System.Net.Http.Json;
using SampleApp.SampleParent.Application;
using SampleApp.SampleParent.Endpoints;

namespace SampleApp.Tests.SampleParent;

public class SampleItemEndpointTests
{
    private const string Route = SampleItemEndpoints.Route;

    [Fact]
    public async Task Cria_renomeia_consulta_lista_e_remove()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        var id = await TestApi.CreateAsync(client, Route, new { name = "Exemplo" });
        Assert.Equal(HttpStatusCode.NoContent, (await client.PutAsJsonAsync($"{Route}/{id}", new { name = "Outro" })).StatusCode);

        var found = await client.GetFromJsonAsync<SampleItemResponse>($"{Route}/{id}");
        Assert.Equal("Outro", found!.Name);
        var listed = await client.GetFromJsonAsync<List<SampleItemResponse>>($"{Route}?limit=10");
        Assert.Equal(id, Assert.Single(listed!).Id);

        Assert.Equal(HttpStatusCode.NoContent, (await client.DeleteAsync($"{Route}/{id}")).StatusCode);
        await TestApi.AssertProblemAsync(await client.GetAsync($"{Route}/{id}"), HttpStatusCode.NotFound);
    }

    [Fact]
    public async Task Responde_400_422_e_404()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var id = await TestApi.CreateAsync(client, Route, new { name = "Exemplo" });

        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync(Route, new { }), HttpStatusCode.BadRequest);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync(Route, new { name = " " }), HttpStatusCode.UnprocessableEntity);
        await TestApi.AssertProblemAsync(await client.PutAsJsonAsync($"{Route}/{id}", new { name = "" }), HttpStatusCode.UnprocessableEntity);
        await TestApi.AssertProblemAsync(await client.PutAsJsonAsync($"{Route}/{Guid.NewGuid()}", new { name = "Outro" }), HttpStatusCode.NotFound);
        await TestApi.AssertProblemAsync(await client.DeleteAsync($"{Route}/{Guid.NewGuid()}"), HttpStatusCode.NotFound);
    }
}
