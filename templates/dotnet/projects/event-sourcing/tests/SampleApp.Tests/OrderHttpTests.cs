using System.Net;
using System.Net.Http.Json;
using System.Text;
using SampleApp.Application.Common;
using SampleApp.Application.EventSourcing;
using SampleApp.Application.Orders;
using SampleApp.Domain.Common;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.TestHost;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Tests;

public class OrderHttpTests(ApiFactory factory) : IClassFixture<ApiFactory>
{
    private sealed record Created(Guid Id);

    private sealed record StreamEvent(long Version, string Type);

    private sealed class ConflictingEventStore : IEventStore
    {
        public Task<IReadOnlyList<StoredEvent>> ReadStreamAsync(Guid streamId, CancellationToken cancellationToken) =>
            Task.FromResult<IReadOnlyList<StoredEvent>>([]);

        public Task AppendAsync(Guid streamId, long expectedVersion, IReadOnlyList<IDomainEvent> events, CancellationToken cancellationToken) =>
            throw new ConcurrencyException(streamId);
    }

    private readonly HttpClient _client = factory.CreateClient();

    private async Task<Guid> CreateOrder(string customer)
    {
        var response = await _client.PostAsJsonAsync("/orders", new { customerName = customer });
        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        return (await response.Content.ReadFromJsonAsync<Created>())!.Id;
    }

    [Fact]
    public async Task Health_responde()
    {
        var response = await _client.GetAsync("/health");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
    }

    [Fact]
    public async Task Fluxo_completo_pela_API()
    {
        var id = await CreateOrder("Ana");

        var added = await _client.PostAsJsonAsync($"/orders/{id}/items", new { productName = "Caneta", quantity = 2, unitPrice = 1.5m });
        Assert.Equal(HttpStatusCode.NoContent, added.StatusCode);
        var placed = await _client.PostAsync($"/orders/{id}/place", null);
        Assert.Equal(HttpStatusCode.NoContent, placed.StatusCode);

        var order = await _client.GetFromJsonAsync<OrderSummary>($"/orders/{id}");
        Assert.Equal("Placed", order!.Status);
        Assert.Equal(3m, order.Total);
        Assert.Contains(await _client.GetFromJsonAsync<List<OrderSummary>>("/orders?limit=200") ?? [], o => o.Id == id);

        var events = (await _client.GetFromJsonAsync<List<StreamEvent>>($"/orders/{id}/events"))!;
        Assert.Equal(["OrderCreated", "ItemAdded", "OrderPlaced"], events.Select(e => e.Type));
        Assert.Equal([1L, 2L, 3L], events.Select(e => e.Version));
    }

    [Fact]
    public async Task Regra_violada_vira_422()
    {
        var response = await _client.PostAsJsonAsync("/orders", new { customerName = " " });

        Assert.Equal(HttpStatusCode.UnprocessableEntity, response.StatusCode);
        Assert.Equal("application/problem+json", response.Content.Headers.ContentType?.MediaType);
        var problem = await response.Content.ReadFromJsonAsync<ProblemDetails>();
        Assert.Equal("O nome do cliente é obrigatório.", problem!.Detail);
    }

    [Fact]
    public async Task Item_com_moeda_diferente_vira_422()
    {
        var id = await CreateOrder("Ana");
        await _client.PostAsJsonAsync($"/orders/{id}/items", new { productName = "Caneta", quantity = 1, unitPrice = 1m });

        var response = await _client.PostAsJsonAsync($"/orders/{id}/items", new { productName = "Pen", quantity = 1, unitPrice = 1m, currency = "USD" });

        Assert.Equal(HttpStatusCode.UnprocessableEntity, response.StatusCode);
    }

    [Fact]
    public async Task Pedido_inexistente_vira_404()
    {
        Assert.Equal(HttpStatusCode.NotFound, (await _client.GetAsync($"/orders/{Guid.NewGuid()}")).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await _client.GetAsync($"/orders/{Guid.NewGuid()}/events")).StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, (await _client.PostAsync($"/orders/{Guid.NewGuid()}/place", null)).StatusCode);
    }

    [Fact]
    public async Task Corpo_invalido_vira_400()
    {
        var response = await _client.PostAsync("/orders", new StringContent("{", Encoding.UTF8, "application/json"));

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    [Fact]
    public async Task Conflito_de_versao_vira_409()
    {
        var client = factory
            .WithWebHostBuilder(builder => builder.ConfigureTestServices(services => services.AddSingleton<IEventStore, ConflictingEventStore>()))
            .CreateClient();

        var response = await client.PostAsJsonAsync("/orders", new { customerName = "Ana" });

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
    }
}
