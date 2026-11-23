using System.Net;
using System.Net.Http.Json;
using SampleApp.Catalog.Application;

namespace SampleApp.Tests.Catalog;

public class ProductEndpointTests
{
    [Fact]
    public async Task Creates_gets_and_lists_products()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        var pen = await TestApi.CreateProductAsync(client, "Pen", 1.50m);
        var notebook = await TestApi.CreateProductAsync(client, "Notebook", 10m, "usd");

        var found = await client.GetFromJsonAsync<ProductResponse>($"/catalog/products/{notebook}");
        Assert.Equal("Notebook", found!.Name);
        Assert.Equal(10m, found.Price);
        Assert.Equal("USD", found.Currency);
        Assert.Equal(0, found.SoldQuantity);

        var listed = await client.GetFromJsonAsync<List<ProductResponse>>("/catalog/products?limit=1");
        Assert.Equal(notebook, Assert.Single(listed!).Id);
        var all = await client.GetFromJsonAsync<List<ProductResponse>>("/catalog/products");
        Assert.Contains(all!, product => product is { Name: "Pen", Currency: "BRL" } && product.Id == pen);
    }

    [Fact]
    public async Task Answers_400_422_and_404()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/catalog/products", new { name = "Pen" }), HttpStatusCode.BadRequest);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/catalog/products", new { price = 1m }), HttpStatusCode.BadRequest);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/catalog/products", new { name = " ", price = 1m }), HttpStatusCode.UnprocessableEntity);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/catalog/products", new { name = "Pen", price = -1m }), HttpStatusCode.UnprocessableEntity);
        await TestApi.AssertProblemAsync(
            await client.PostAsJsonAsync("/catalog/products", new { name = "Pen", price = 1m, currency = "REAL" }),
            HttpStatusCode.UnprocessableEntity);
        await TestApi.AssertProblemAsync(await client.GetAsync($"/catalog/products/{Guid.NewGuid()}"), HttpStatusCode.NotFound);
    }
}
