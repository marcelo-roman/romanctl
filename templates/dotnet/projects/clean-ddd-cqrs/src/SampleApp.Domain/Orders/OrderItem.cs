using SampleApp.Domain.Common;

namespace SampleApp.Domain.Orders;

public sealed class OrderItem
{
    public Guid Id { get; private set; }
    public string ProductName { get; private set; } = default!;
    public int Quantity { get; private set; }
    public Money UnitPrice { get; private set; } = default!;

    public Money Subtotal => UnitPrice.Multiply(Quantity);

    private OrderItem() { }

    internal OrderItem(string productName, int quantity, Money unitPrice)
    {
        if (string.IsNullOrWhiteSpace(productName))
            throw new DomainException("O nome do produto é obrigatório.");
        if (quantity <= 0)
            throw new DomainException("A quantidade deve ser maior que zero.");

        Id = Guid.NewGuid();
        ProductName = productName.Trim();
        Quantity = quantity;
        UnitPrice = unitPrice;
    }
}
