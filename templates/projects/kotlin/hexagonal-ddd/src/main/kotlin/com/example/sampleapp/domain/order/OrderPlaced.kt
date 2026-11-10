package com.example.sampleapp.domain.order

import com.example.sampleapp.domain.shared.DomainEvent
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

data class OrderPlaced(
    val orderId: UUID,
    val total: BigDecimal,
    val currency: String,
    override val occurredAt: Instant,
) : DomainEvent {
    override val name: String = "order.placed"
}
