namespace SampleApp.SampleItems.Contracts;

public interface ISampleItemsApi
{
    Task<SampleItemSnapshot?> FindSampleItemAsync(Guid id, CancellationToken cancellationToken);
}
