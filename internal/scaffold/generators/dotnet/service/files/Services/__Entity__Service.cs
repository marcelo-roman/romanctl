namespace {{.Name}}.Services;

public interface I{{.Entity}}Service
{
    Task<string> GetAsync(string id, CancellationToken cancellationToken);
}

public sealed class {{.Entity}}Service : I{{.Entity}}Service
{
    public Task<string> GetAsync(string id, CancellationToken cancellationToken) => Task.FromResult(id);
}
