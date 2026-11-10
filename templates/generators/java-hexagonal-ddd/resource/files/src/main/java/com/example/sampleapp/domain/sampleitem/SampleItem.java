package com.example.sampleapp.domain.sampleitem;

import com.example.sampleapp.domain.shared.BusinessRuleException;
import java.time.Instant;
import java.util.UUID;

public final class SampleItem {

    private final UUID id;
    private final Instant createdAt;
    private String name;

    private SampleItem(UUID id, String name, Instant createdAt) {
        this.id = id;
        this.name = name;
        this.createdAt = createdAt;
    }

    public static SampleItem create(String name, Instant now) {
        SampleItem sampleItem = new SampleItem(UUID.randomUUID(), null, now);
        sampleItem.rename(name);
        return sampleItem;
    }

    public static SampleItem restore(UUID id, String name, Instant createdAt) {
        return new SampleItem(id, name, createdAt);
    }

    public UUID id() {
        return id;
    }

    public String name() {
        return name;
    }

    public Instant createdAt() {
        return createdAt;
    }

    public void rename(String name) {
        if (name == null || name.isBlank()) {
            throw new BusinessRuleException("O nome é obrigatório.");
        }
        this.name = name.strip();
    }
}
