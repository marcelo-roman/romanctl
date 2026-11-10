using System.Net;
using System.Net.Http.Json;
using SampleApp.Catalog.Application;
using SampleApp.Ordering.Application;

namespace SampleApp.Tests.Modules;

public class OrderPlacedFlowTests
{
    [Fact]
    public async Task Pedido_realizado_chega_ao_catalogo_pelo_evento_de_integracao()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var pen = await TestApi.CreateProductAsync(client, "Caneta", 1.50m);
        var notebook = await TestApi.CreateProductAsync(client, "Caderno", 10m);
        var id = await TestApi.CreateOrderAsync(client);
        Assert.Equal(HttpStatusCode.NoContent, await TestApi.AddItemAsync(client, id, pen, 2));
        Assert.Equal(HttpStatusCode.NoContent, await TestApi.AddItemAsync(client, id, notebook, 1));
        Assert.Equal(HttpStatusCode.NoContent, await TestApi.AddItemAsync(client, id, pen, 3));

        Assert.Equal(HttpStatusCode.NoContent, (await client.PostAsync($"/orders/{id}/place", null)).StatusCode);

        var order = await client.GetFromJsonAsync<OrderResponse>($"/orders/{id}");
        Assert.Equal("Placed", order!.Status);
        Assert.Equal(17.50m, order.Total);
        var soldPens = await Eventually.MatchAsync(
            () => client.GetFromJsonAsync<ProductResponse>($"/catalog/products/{pen}"),
            product => product!.SoldQuantity > 0);
        Assert.Equal(5, soldPens!.SoldQuantity);
        var soldNotebooks = await client.GetFromJsonAsync<ProductResponse>($"/catalog/products/{notebook}");
        Assert.Equal(1, soldNotebooks!.SoldQuantity);
    }
}
