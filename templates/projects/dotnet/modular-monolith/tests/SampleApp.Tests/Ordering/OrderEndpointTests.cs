using System.Net;
using System.Net.Http.Json;
using System.Text;
using SampleApp.Ordering.Application;

namespace SampleApp.Tests.Ordering;

public class OrderEndpointTests
{
    [Fact]
    public async Task Health_responde_ok()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        var response = await client.GetAsync("/health");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
    }

    [Fact]
    public async Task Item_vem_do_catalogo_com_nome_e_preco_do_produto()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var pen = await TestApi.CreateProductAsync(client, "Caneta", 1.50m);
        var notebook = await TestApi.CreateProductAsync(client, "Caderno", 10m);
        var id = await TestApi.CreateOrderAsync(client);

        Assert.Equal(HttpStatusCode.NoContent, await TestApi.AddItemAsync(client, id, pen, 2));
        Assert.Equal(HttpStatusCode.NoContent, await TestApi.AddItemAsync(client, id, notebook, 1));

        var order = await client.GetFromJsonAsync<OrderResponse>($"/orders/{id}");
        Assert.Equal("Ana", order!.CustomerName);
        Assert.Equal("Draft", order.Status);
        Assert.Equal(13m, order.Total);
        Assert.Equal("BRL", order.Currency);
        Assert.Contains(order.Items, item => item is { ProductName: "Caneta", Quantity: 2, UnitPrice: 1.50m, Subtotal: 3m } && item.ProductId == pen);
    }

    [Fact]
    public async Task Lista_os_mais_recentes_primeiro_respeitando_o_limite()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var first = await TestApi.CreateOrderAsync(client, "Ana");
        var second = await TestApi.CreateOrderAsync(client, "Bia");
        var third = await TestApi.CreateOrderAsync(client, "Caio");

        var listed = await client.GetFromJsonAsync<List<OrderResponse>>("/orders?limit=2");

        Assert.Equal([third, second], listed!.Select(order => order.Id));
        Assert.DoesNotContain(listed!, order => order.Id == first);
    }

    [Fact]
    public async Task Regra_de_negocio_violada_vira_422()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var pen = await TestApi.CreateProductAsync(client, "Caneta", 1m);
        var dollar = await TestApi.CreateProductAsync(client, "Pen", 1m, "USD");
        var id = await TestApi.CreateOrderAsync(client);

        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/orders", new { customerName = " " }), HttpStatusCode.UnprocessableEntity);
        await TestApi.AssertProblemAsync(await client.PostAsync($"/orders/{id}/place", null), HttpStatusCode.UnprocessableEntity);
        Assert.Equal(HttpStatusCode.UnprocessableEntity, await TestApi.AddItemAsync(client, id, pen, 0));
        Assert.Equal(HttpStatusCode.NoContent, await TestApi.AddItemAsync(client, id, pen, 1));
        Assert.Equal(HttpStatusCode.UnprocessableEntity, await TestApi.AddItemAsync(client, id, dollar, 1));
    }

    [Fact]
    public async Task Pedido_ou_produto_inexistente_vira_404()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var pen = await TestApi.CreateProductAsync(client);
        var id = await TestApi.CreateOrderAsync(client);
        var missing = Guid.NewGuid();

        await TestApi.AssertProblemAsync(await client.GetAsync($"/orders/{missing}"), HttpStatusCode.NotFound);
        await TestApi.AssertProblemAsync(await client.PostAsync($"/orders/{missing}/place", null), HttpStatusCode.NotFound);
        await TestApi.AssertProblemAsync(
            await client.PostAsJsonAsync($"/orders/{missing}/items", new { productId = pen, quantity = 1 }),
            HttpStatusCode.NotFound);
        await TestApi.AssertProblemAsync(
            await client.PostAsJsonAsync($"/orders/{id}/items", new { productId = missing, quantity = 1 }),
            HttpStatusCode.NotFound);
    }

    [Fact]
    public async Task Corpo_invalido_vira_400()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var id = await TestApi.CreateOrderAsync(client);

        await TestApi.AssertProblemAsync(
            await client.PostAsync("/orders", new StringContent("{", Encoding.UTF8, "application/json")),
            HttpStatusCode.BadRequest);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/orders", new { }), HttpStatusCode.BadRequest);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync("/orders", new { customerName = (string?)null }), HttpStatusCode.BadRequest);
        await TestApi.AssertProblemAsync(
            await client.PostAsJsonAsync($"/orders/{id}/items", new { productId = Guid.NewGuid(), quantity = "muitas" }),
            HttpStatusCode.BadRequest);
        await TestApi.AssertProblemAsync(await client.PostAsJsonAsync($"/orders/{id}/items", new { productId = Guid.NewGuid() }), HttpStatusCode.BadRequest);
    }
}
