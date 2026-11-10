using System.Net;
using System.Net.Http.Json;
using SampleApp.Shared.Web;
using Microsoft.AspNetCore.Mvc;

namespace SampleApp.Tests;

public static class TestApi
{
    public static async Task<Guid> CreateAsync(HttpClient client, string route, object body)
    {
        var response = await client.PostAsJsonAsync(route, body);
        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        var created = await response.Content.ReadFromJsonAsync<CreatedResponse>();
        Assert.Equal($"{route}/{created!.Id}", response.Headers.Location?.OriginalString);
        return created.Id;
    }

    public static Task<Guid> CreateProductAsync(HttpClient client, string name = "Caneta", decimal price = 1.50m, string? currency = null) =>
        CreateAsync(client, "/catalog/products", new { name, price, currency });

    public static Task<Guid> CreateOrderAsync(HttpClient client, string customerName = "Ana") =>
        CreateAsync(client, "/orders", new { customerName });

    public static async Task<HttpStatusCode> AddItemAsync(HttpClient client, Guid orderId, Guid productId, int quantity) =>
        (await client.PostAsJsonAsync($"/orders/{orderId}/items", new { productId, quantity })).StatusCode;

    public static async Task AssertProblemAsync(HttpResponseMessage response, HttpStatusCode status)
    {
        Assert.Equal(status, response.StatusCode);
        Assert.Equal("application/problem+json", response.Content.Headers.ContentType?.MediaType);
        var problem = await response.Content.ReadFromJsonAsync<ProblemDetails>();
        Assert.Equal((int)status, problem!.Status);
    }
}
