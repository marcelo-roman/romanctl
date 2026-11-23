using System.Net;
using System.Net.Http.Json;
using System.Text;
using SampleApp.Api.Features.Orders;
using Microsoft.AspNetCore.Mvc;

namespace SampleApp.Tests.Orders;

public class OrderEndpointTests
{
    private static async Task<Guid> CreateOrderAsync(HttpClient client, string customerName = "Ana")
    {
        var response = await client.PostAsJsonAsync("/orders", new { customerName });
        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var created = await response.Content.ReadFromJsonAsync<CreateOrder.Response>();
        Assert.Equal($"/orders/{created!.Id}", response.Headers.Location?.OriginalString);
        return created.Id;
    }

    private static async Task<HttpStatusCode> AddItemAsync(HttpClient client, Guid id, object item) =>
        (await client.PostAsJsonAsync($"/orders/{id}/items", item)).StatusCode;

    private static async Task AssertProblemAsync(HttpResponseMessage response, HttpStatusCode status)
    {
        Assert.Equal(status, response.StatusCode);
        Assert.Equal("application/problem+json", response.Content.Headers.ContentType?.MediaType);
        var problem = await response.Content.ReadFromJsonAsync<ProblemDetails>();
        Assert.Equal((int)status, problem!.Status);
    }

    [Fact]
    public async Task Health_answers_ok()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        var response = await client.GetAsync("/health");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
    }

    [Fact]
    public async Task Full_order_flow()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();

        var id = await CreateOrderAsync(client);
        Assert.Equal(HttpStatusCode.NoContent, await AddItemAsync(client, id, new { productName = "Pen", quantity = 2, unitPrice = 1.50m }));
        Assert.Equal(HttpStatusCode.NoContent, await AddItemAsync(client, id, new { productName = "Notebook", quantity = 1, unitPrice = 10m, currency = "brl" }));
        Assert.Equal(HttpStatusCode.NoContent, (await client.PostAsync($"/orders/{id}/place", null)).StatusCode);

        var order = await client.GetFromJsonAsync<GetOrder.Response>($"/orders/{id}");
        Assert.Equal("Ana", order!.CustomerName);
        Assert.Equal("Placed", order.Status);
        Assert.Equal(13m, order.Total);
        Assert.Equal("BRL", order.Currency);
        Assert.Equal(2, order.Items.Count);
        Assert.Contains(order.Items, item => item is { ProductName: "Pen", Quantity: 2, Subtotal: 3m });

        var listed = await client.GetFromJsonAsync<List<ListOrders.Response>>("/orders");
        Assert.Equal(id, Assert.Single(listed!).Id);
    }

    [Fact]
    public async Task Lists_most_recent_first_respecting_the_limit()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var first = await CreateOrderAsync(client, "Ana");
        var second = await CreateOrderAsync(client, "Bob");
        var third = await CreateOrderAsync(client, "Carl");

        var listed = await client.GetFromJsonAsync<List<ListOrders.Response>>("/orders?limit=2");

        Assert.Equal([third, second], listed!.Select(order => order.Id));
        Assert.DoesNotContain(listed!, order => order.Id == first);
    }

    [Fact]
    public async Task Business_rule_violation_becomes_422()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var id = await CreateOrderAsync(client);

        await AssertProblemAsync(await client.PostAsJsonAsync("/orders", new { customerName = " " }), HttpStatusCode.UnprocessableEntity);
        await AssertProblemAsync(await client.PostAsync($"/orders/{id}/place", null), HttpStatusCode.UnprocessableEntity);
        Assert.Equal(HttpStatusCode.UnprocessableEntity, await AddItemAsync(client, id, new { productName = "Pen", quantity = 0, unitPrice = 1m }));
        Assert.Equal(HttpStatusCode.UnprocessableEntity, await AddItemAsync(client, id, new { productName = "Pen", quantity = 1, unitPrice = -1m }));
        Assert.Equal(HttpStatusCode.UnprocessableEntity, await AddItemAsync(client, id, new { productName = "Pen", quantity = 1, unitPrice = 1m, currency = "REAL" }));
        Assert.Equal(HttpStatusCode.NoContent, await AddItemAsync(client, id, new { productName = "Pen", quantity = 1, unitPrice = 1m }));
        Assert.Equal(HttpStatusCode.UnprocessableEntity, await AddItemAsync(client, id, new { productName = "Imported pen", quantity = 1, unitPrice = 1m, currency = "USD" }));
    }

    [Fact]
    public async Task Unknown_order_becomes_404()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var id = Guid.NewGuid();

        await AssertProblemAsync(await client.GetAsync($"/orders/{id}"), HttpStatusCode.NotFound);
        await AssertProblemAsync(await client.PostAsync($"/orders/{id}/place", null), HttpStatusCode.NotFound);
        await AssertProblemAsync(
            await client.PostAsJsonAsync($"/orders/{id}/items", new { productName = "Pen", quantity = 1, unitPrice = 1m }),
            HttpStatusCode.NotFound);
    }

    [Fact]
    public async Task Invalid_body_becomes_400()
    {
        await using var factory = new ApiFactory();
        using var client = factory.CreateClient();
        var id = await CreateOrderAsync(client);

        await AssertProblemAsync(
            await client.PostAsync("/orders", new StringContent("{", Encoding.UTF8, "application/json")),
            HttpStatusCode.BadRequest);
        await AssertProblemAsync(await client.PostAsJsonAsync("/orders", new { }), HttpStatusCode.BadRequest);
        await AssertProblemAsync(
            await client.PostAsJsonAsync($"/orders/{id}/items", new { productName = "Pen", quantity = "many", unitPrice = 1m }),
            HttpStatusCode.BadRequest);
        await AssertProblemAsync(await client.PostAsJsonAsync($"/orders/{id}/items", new { productName = "Pen" }), HttpStatusCode.BadRequest);
    }
}
