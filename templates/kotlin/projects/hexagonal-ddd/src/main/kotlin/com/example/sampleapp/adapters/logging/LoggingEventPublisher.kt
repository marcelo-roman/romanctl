package com.example.sampleapp.adapters.logging

import com.example.sampleapp.application.shared.EventPublisher
import com.example.sampleapp.domain.shared.DomainEvent
import org.slf4j.LoggerFactory

class LoggingEventPublisher : EventPublisher {
    private val log = LoggerFactory.getLogger(LoggingEventPublisher::class.java)

    override fun publish(events: List<DomainEvent>) {
        events.forEach { log.info("domain event {}: {}", it.name, it) }
    }
}
