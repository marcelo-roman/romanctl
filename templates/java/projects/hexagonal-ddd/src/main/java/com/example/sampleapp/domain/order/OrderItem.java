package com.example.sampleapp.domain.order;

import com.example.sampleapp.domain.shared.BusinessRuleException;
import java.util.UUID;

public record OrderItem(UUID id, String productName, int quantity, Money unitPrice) {

    public OrderItem {
        if (productName == null || productName.isBlank()) {
            throw new BusinessRuleException("Product name is required.");
        }
        if (quantity <= 0) {
            throw new BusinessRuleException("Quantity must be greater than zero.");
        }
        productName = productName.strip();
    }

    static OrderItem create(String productName, int quantity, Money unitPrice) {
        return new OrderItem(UUID.randomUUID(), productName, quantity, unitPrice);
    }

    public Money subtotal() {
        return unitPrice.multiply(quantity);
    }
}
