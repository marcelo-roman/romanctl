using System.Net;
using System.Net.Http.Json;
using SampleApp.Catalog.Application;

namespace SampleApp.Tests.Catalog;

public class ProductEndpointTests
{
    [Fact]
    public async Task Cria_consulta_e_lista_produtos()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        var pen = await TestApi.CreateProductAsync(client, "Caneta", 1.50m);
        var notebook = await TestApi.CreateProductAsync(client, "Caderno", 10m, "usd");

        var found = await client.GetFromJsonAsync<ProductResponse>($"/catalog/products/{notebook}");
        Assert.Equal("Caderno", found!.Name);
        Assert.Equal(10m, found.Price);
        Assert.Equal("USD", found.Currency);
        Assert.Equal(0, found.SoldQuantity);

        var listed = await client.GetFromJsonAsync<List<ProductResponse>>("/catalog/products?limit=1");
        Assert.Equal(notebook, Assert.Single(listed!).Id);
        var all = await client.GetFromJsonAsync<List<ProductResponse>>("/catalog/products");
        Assert.Contains(all!, product => product is { Name: "Caneta", Currency: "BRL" } && product.Id == pen);
    }

    [Fact]
    public async Task Responde_400_422_e_404()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/catalog/products", new { name = "Caneta" }), HttpStatusCode.BadRequest);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/catalog/products", new { price = 1m }), HttpStatusCode.BadRequest);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/catalog/products", new { name = " ", price = 1m }), HttpStatusCode.UnprocessableEntity);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/catalog/products", new { name = "Caneta", price = -1m }), HttpStatusCode.UnprocessableEntity);
        await TestApi.AssertProblemAsync(
            await client.PostAsJsonAsync("/catalog/products", new { name = "Caneta", price = 1m, currency = "REAL" }),
            HttpStatusCode.UnprocessableEntity);
        await TestApi.AssertProblemAsync(await client.GetAsync($"/catalog/products/{Guid.NewGuid()}"), HttpStatusCode.NotFound);
    }
}
