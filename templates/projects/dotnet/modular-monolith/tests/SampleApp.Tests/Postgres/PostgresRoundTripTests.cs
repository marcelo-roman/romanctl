using System.Net;
using System.Net.Http.Json;
using SampleApp.Catalog.Application;
using SampleApp.Catalog.Infrastructure;
using SampleApp.Ordering.Application;
using SampleApp.Ordering.Infrastructure;
using SampleApp.Shared.Application;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Tests.Postgres;

[Collection(PostgresCollection.Name)]
public class PostgresRoundTripTests
{
    private static async Task<List<string>> TablesInAsync(WebApplicationFactory<Program> factory, string schema)
    {
        await using var scope = factory.Services.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<CatalogDbContext>();
        return await db.Database
            .SqlQuery<string>($"select table_name as \"Value\" from information_schema.tables where table_schema = {schema}")
            .ToListAsync();
    }

    [PostgresFact]
    public async Task Cada_modulo_tem_as_tabelas_no_proprio_esquema()
    {
        await using var factory = new WebApplicationFactory<Program>();
        using var client = factory.CreateClient();

        var catalog = await TablesInAsync(factory, CatalogDbContext.Schema);
        var ordering = await TablesInAsync(factory, OrderingDbContext.Schema);

        Assert.Contains("products", catalog);
        Assert.Contains("orders", ordering);
        Assert.Contains("order_items", ordering);
        Assert.DoesNotContain("orders", catalog);
        Assert.DoesNotContain("products", ordering);
    }

    [PostgresFact]
    public async Task Pedido_e_venda_vao_e_voltam_do_PostgreSQL()
    {
        await using var factory = new WebApplicationFactory<Program>();
        using var client = factory.CreateClient();

        var pen = await TestApi.CreateProductAsync(client, "Caneta", 1.25m);
        var id = await TestApi.CreateOrderAsync(client);
        Assert.Equal(HttpStatusCode.NoContent, await TestApi.AddItemAsync(client, id, pen, 2));
        Assert.Equal(HttpStatusCode.NoContent, (await client.PostAsync($"/orders/{id}/place", null)).StatusCode);

        var order = await client.GetFromJsonAsync<OrderResponse>($"/orders/{id}");
        Assert.Equal("Placed", order!.Status);
        Assert.Equal(2.50m, order.Total);
        Assert.Equal("BRL", order.Currency);
        Assert.Equal("Caneta", Assert.Single(order.Items).ProductName);
        var listed = await client.GetFromJsonAsync<List<OrderResponse>>($"/orders?limit={ListLimit.Max}");
        Assert.Contains(listed!, recent => recent.Id == id);

        var product = await Eventually.MatchAsync(
            () => client.GetFromJsonAsync<ProductResponse>($"/catalog/products/{pen}"),
            found => found!.SoldQuantity > 0);
        Assert.Equal(2, product!.SoldQuantity);
        Assert.Equal(1.25m, product.Price);
    }
}
