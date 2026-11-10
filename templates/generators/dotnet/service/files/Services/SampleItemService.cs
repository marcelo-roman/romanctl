namespace SampleApp.Services;

public interface ISampleItemService
{
    Task<string> GetAsync(string id, CancellationToken cancellationToken);
}

public sealed class SampleItemService : ISampleItemService
{
    public Task<string> GetAsync(string id, CancellationToken cancellationToken) => Task.FromResult(id);
}
