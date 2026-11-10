using SampleApp.SampleParent.Domain;

namespace SampleApp.SampleParent.Application;

internal sealed record CreateSampleItem(string Name);

internal sealed class CreateSampleItemHandler(ISampleItemRepository sampleItems, TimeProvider clock)
{
    public async Task<Guid> HandleAsync(CreateSampleItem command, CancellationToken cancellationToken)
    {
        var sampleItem = SampleItem.Create(command.Name, clock.GetUtcNow());
        sampleItems.Add(sampleItem);
        await sampleItems.SaveChangesAsync(cancellationToken);
        return sampleItem.Id;
    }
}
