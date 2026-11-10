using SampleApp.Catalog.Contracts;
using SampleApp.Ordering.Domain;
using SampleApp.Shared.Application;
using SampleApp.Shared.Domain;

namespace SampleApp.Ordering.Application;

internal sealed record AddOrderItem(Guid OrderId, Guid ProductId, int Quantity);

internal sealed class AddOrderItemHandler(IOrderRepository orders, ICatalogApi catalog)
{
    public async Task HandleAsync(AddOrderItem command, CancellationToken cancellationToken)
    {
        var order = await orders.FindAsync(command.OrderId, cancellationToken)
            ?? throw new NotFoundException($"Pedido {command.OrderId} não encontrado.");
        var product = await catalog.FindProductAsync(command.ProductId, cancellationToken)
            ?? throw new NotFoundException($"Produto {command.ProductId} não encontrado.");

        order.AddItem(product.Id, product.Name, command.Quantity, new Money(product.Price, product.Currency));
        await orders.SaveChangesAsync(cancellationToken);
    }
}
