package com.example.sampleapp.adapters.memory;

import com.example.sampleapp.application.sampleitems.SampleItemRepository;
import com.example.sampleapp.domain.sampleitem.SampleItem;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Repository;

@Repository
@ConditionalOnProperty(name = "app.persistence", havingValue = "memory", matchIfMissing = true)
public class MemorySampleItemRepository implements SampleItemRepository {

    private final Map<UUID, SampleItem> sampleItems = new ConcurrentHashMap<>();

    @Override
    public void save(SampleItem sampleItem) {
        sampleItems.put(sampleItem.id(), copy(sampleItem));
    }

    @Override
    public Optional<SampleItem> findById(UUID id) {
        return Optional.ofNullable(sampleItems.get(id)).map(MemorySampleItemRepository::copy);
    }

    @Override
    public List<SampleItem> list(int limit) {
        return sampleItems.values().stream()
                .sorted(Comparator.comparing(SampleItem::createdAt).reversed())
                .limit(limit)
                .map(MemorySampleItemRepository::copy)
                .toList();
    }

    @Override
    public boolean delete(UUID id) {
        return sampleItems.remove(id) != null;
    }

    private static SampleItem copy(SampleItem sampleItem) {
        return SampleItem.restore(sampleItem.id(), sampleItem.name(), sampleItem.createdAt());
    }
}
