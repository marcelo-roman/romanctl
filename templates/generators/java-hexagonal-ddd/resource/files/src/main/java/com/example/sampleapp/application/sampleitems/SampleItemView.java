package com.example.sampleapp.application.sampleitems;

import com.example.sampleapp.domain.sampleitem.SampleItem;
import java.time.Instant;
import java.util.UUID;

public record SampleItemView(UUID id, String name, Instant createdAt) {

    static SampleItemView of(SampleItem sampleItem) {
        return new SampleItemView(sampleItem.id(), sampleItem.name(), sampleItem.createdAt());
    }
}
