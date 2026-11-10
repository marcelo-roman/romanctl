using SampleApp.Application;
using SampleApp.Application.Messaging;
using SampleApp.Application.SampleItems.Commands;
using SampleApp.Application.SampleItems.Queries;
using SampleApp.Infrastructure;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace SampleApp.Tests.Postgres;

[Collection(PostgresCollection.Name)]
public class SampleItemMartenTests
{
    private static async Task<T> Send<T>(ServiceProvider provider, Func<IDispatcher, Task<T>> work)
    {
        await using var scope = provider.CreateAsyncScope();
        return await work(scope.ServiceProvider.GetRequiredService<IDispatcher>());
    }

    [PostgresFact]
    public async Task Grava_os_eventos_e_atualiza_a_projecao()
    {
        var configuration = new ConfigurationBuilder().AddEnvironmentVariables().Build();
        await using var provider = new ServiceCollection().AddLogging().AddApplication().AddInfrastructure(configuration).BuildServiceProvider();

        var id = await Send(provider, dispatcher => dispatcher.SendAsync(new CreateSampleItem("Primeiro")));
        await Send(provider, dispatcher => dispatcher.SendAsync(new RenameSampleItem(id, "Segundo")));

        var summary = await Send(provider, dispatcher => dispatcher.QueryAsync(new GetSampleItemById(id)));
        Assert.Equal("Segundo", summary.Name);
        Assert.Equal(1, summary.Renames);
        var events = await Send(provider, dispatcher => dispatcher.QueryAsync(new GetSampleItemEvents(id)));
        Assert.Equal(["SampleItemCreated", "SampleItemRenamed"], events.Select(e => e.Type));
    }
}
