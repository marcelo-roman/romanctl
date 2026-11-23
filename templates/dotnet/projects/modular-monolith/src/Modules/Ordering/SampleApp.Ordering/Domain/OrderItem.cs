using SampleApp.Shared.Domain;

namespace SampleApp.Ordering.Domain;

internal sealed class OrderItem
{
    public Guid Id { get; private set; }
    public Guid ProductId { get; private set; }
    public string ProductName { get; private set; } = default!;
    public int Quantity { get; private set; }
    public Money UnitPrice { get; private set; } = default!;

    public Money Subtotal => UnitPrice.Multiply(Quantity);

    private OrderItem() { }

    internal OrderItem(Guid productId, string productName, int quantity, Money unitPrice)
    {
        if (productId == Guid.Empty)
            throw new DomainException("Product is required.");
        if (string.IsNullOrWhiteSpace(productName))
            throw new DomainException("Product name is required.");
        if (quantity <= 0)
            throw new DomainException("Quantity must be greater than zero.");

        Id = Guid.NewGuid();
        ProductId = productId;
        ProductName = productName.Trim();
        Quantity = quantity;
        UnitPrice = unitPrice;
    }
}
