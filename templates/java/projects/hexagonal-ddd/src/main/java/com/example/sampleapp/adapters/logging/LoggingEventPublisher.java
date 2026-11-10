package com.example.sampleapp.adapters.logging;

import com.example.sampleapp.application.shared.EventPublisher;
import com.example.sampleapp.domain.shared.DomainEvent;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

@Component
public class LoggingEventPublisher implements EventPublisher {

    private static final Logger LOG = LoggerFactory.getLogger(LoggingEventPublisher.class);

    @Override
    public void publish(List<DomainEvent> events) {
        events.forEach(event -> LOG.info("domain event {}: {}", event.name(), event));
    }
}
