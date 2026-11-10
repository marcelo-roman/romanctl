package com.example.sampleapp.domain.order;

import com.example.sampleapp.domain.shared.DomainEvent;
import java.math.BigDecimal;
import java.time.Instant;
import java.util.UUID;

public record OrderPlaced(UUID orderId, BigDecimal total, String currency, Instant occurredAt) implements DomainEvent {

    @Override
    public String name() {
        return "order.placed";
    }
}
