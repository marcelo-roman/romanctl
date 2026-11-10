namespace SampleApp.Domain.Orders;

public sealed record OrderLine(string ProductName, int Quantity, Money UnitPrice)
{
    public Money Subtotal => UnitPrice.Multiply(Quantity);
}
