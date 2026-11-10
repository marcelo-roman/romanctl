using SampleApp.Catalog.Domain;
using SampleApp.Ordering.Contracts;
using SampleApp.Shared.Messaging;
using Microsoft.Extensions.Logging;

namespace SampleApp.Catalog.Application;

internal sealed class RecordSalesOnOrderPlaced(IProductRepository products, ILogger<RecordSalesOnOrderPlaced> logger)
    : IIntegrationEventHandler<OrderPlaced>
{
    public async Task HandleAsync(OrderPlaced integrationEvent, CancellationToken cancellationToken)
    {
        var ids = integrationEvent.Lines.Select(line => line.ProductId).ToHashSet();
        var sold = (await products.FindManyAsync(ids, cancellationToken)).ToDictionary(product => product.Id);

        foreach (var line in integrationEvent.Lines)
        {
            if (!sold.TryGetValue(line.ProductId, out var product))
            {
                logger.LogWarning("Pedido {OrderId}: produto {ProductId} não está mais no catálogo.", integrationEvent.OrderId, line.ProductId);
                continue;
            }
            product.RecordSale(line.Quantity);
        }

        await products.SaveChangesAsync(cancellationToken);
        logger.LogInformation("Pedido {OrderId}: vendas registradas no catálogo.", integrationEvent.OrderId);
    }
}
