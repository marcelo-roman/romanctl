using System.Net;
using System.Net.Http.Json;
using SampleApp.Api.Features.Orders;
using Microsoft.AspNetCore.Mvc.Testing;

namespace SampleApp.Tests.Postgres;

public class PostgresRoundTripTests
{
    [PostgresFact]
    public async Task Order_round_trips_through_PostgreSQL()
    {
        await using var factory = new WebApplicationFactory<Program>();
        using var client = factory.CreateClient();

        var created = await client.PostAsJsonAsync("/orders", new { customerName = "Ana" });
        Assert.Equal(HttpStatusCode.Created, created.StatusCode);
        var id = (await created.Content.ReadFromJsonAsync<CreateOrder.Response>())!.Id;
        var item = await client.PostAsJsonAsync($"/orders/{id}/items", new { productName = "Pen", quantity = 2, unitPrice = 1.25m });
        Assert.Equal(HttpStatusCode.NoContent, item.StatusCode);
        Assert.Equal(HttpStatusCode.NoContent, (await client.PostAsync($"/orders/{id}/place", null)).StatusCode);

        var order = await client.GetFromJsonAsync<GetOrder.Response>($"/orders/{id}");

        Assert.Equal("Placed", order!.Status);
        Assert.Equal(2.50m, order.Total);
        Assert.Equal("BRL", order.Currency);
        Assert.Equal("Pen", Assert.Single(order.Items).ProductName);
        var listed = await client.GetFromJsonAsync<List<ListOrders.Response>>($"/orders?limit={ListOrders.MaxLimit}");
        Assert.Contains(listed!, order => order.Id == id);
    }
}
