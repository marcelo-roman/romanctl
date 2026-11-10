using SampleApp.Catalog.Domain;

namespace SampleApp.Catalog.Application;

internal sealed record ProductResponse(Guid Id, string Name, decimal Price, string Currency, int SoldQuantity, DateTimeOffset CreatedAt)
{
    public static ProductResponse From(Product product) =>
        new(product.Id, product.Name, product.Price.Amount, product.Price.Currency, product.SoldQuantity, product.CreatedAt);
}
