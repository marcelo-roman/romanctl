package com.example.sampleapp.domain.order

import com.example.sampleapp.domain.shared.BusinessRuleException
import java.util.UUID

class OrderItem(val id: UUID, productName: String, val quantity: Int, val unitPrice: Money) {
    val productName: String

    init {
        if (productName.isBlank()) {
            throw BusinessRuleException("O nome do produto é obrigatório.")
        }
        if (quantity <= 0) {
            throw BusinessRuleException("A quantidade deve ser maior que zero.")
        }
        this.productName = productName.trim()
    }

    val subtotal: Money
        get() = unitPrice * quantity

    companion object {
        internal fun create(productName: String, quantity: Int, unitPrice: Money): OrderItem =
            OrderItem(UUID.randomUUID(), productName, quantity, unitPrice)
    }
}
