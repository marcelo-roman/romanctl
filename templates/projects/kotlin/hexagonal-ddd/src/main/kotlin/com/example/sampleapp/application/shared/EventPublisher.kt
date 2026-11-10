package com.example.sampleapp.application.shared

import com.example.sampleapp.domain.shared.DomainEvent

interface EventPublisher {
    fun publish(events: List<DomainEvent>)
}
