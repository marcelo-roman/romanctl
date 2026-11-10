package com.example.sampleapp.application.shared;

import com.example.sampleapp.domain.shared.DomainEvent;
import java.util.List;

public interface EventPublisher {

    void publish(List<DomainEvent> events);
}
