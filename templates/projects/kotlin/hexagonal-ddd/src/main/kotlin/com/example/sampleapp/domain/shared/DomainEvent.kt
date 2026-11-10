package com.example.sampleapp.domain.shared

import java.time.Instant

interface DomainEvent {
    val name: String
    val occurredAt: Instant
}
