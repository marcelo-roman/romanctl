using SampleApp.Application.EventSourcing;
using SampleApp.Domain.SampleItems;

namespace SampleApp.Application.SampleItems;

public sealed class SampleItemSummaryProjection : Projection<SampleItemSummary>
{
    public SampleItemSummaryProjection()
    {
        Creates<SampleItemCreated>((id, created) => new SampleItemSummary
        {
            Id = id,
            Name = created.Name,
            CreatedAt = created.CreatedAt,
        });

        Updates<SampleItemRenamed>((summary, renamed) => summary with
        {
            Name = renamed.Name,
            Renames = summary.Renames + 1,
        });
    }
}
