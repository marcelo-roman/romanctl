package com.example.sampleapp.domain.shared;

import java.time.Instant;

public interface DomainEvent {

    String name();

    Instant occurredAt();
}
