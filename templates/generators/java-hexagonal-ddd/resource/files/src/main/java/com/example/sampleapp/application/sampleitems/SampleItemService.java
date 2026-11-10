package com.example.sampleapp.application.sampleitems;

import com.example.sampleapp.application.shared.ListLimit;
import com.example.sampleapp.application.shared.NotFoundException;
import com.example.sampleapp.domain.sampleitem.SampleItem;
import java.time.Clock;
import java.util.List;
import java.util.UUID;

public class SampleItemService {

    private final SampleItemRepository repository;
    private final Clock clock;

    public SampleItemService(SampleItemRepository repository, Clock clock) {
        this.repository = repository;
        this.clock = clock;
    }

    public UUID create(String name) {
        SampleItem sampleItem = SampleItem.create(name, clock.instant());
        repository.save(sampleItem);
        return sampleItem.id();
    }

    public void rename(UUID id, String name) {
        SampleItem sampleItem = load(id);
        sampleItem.rename(name);
        repository.save(sampleItem);
    }

    public SampleItemView get(UUID id) {
        return SampleItemView.of(load(id));
    }

    public List<SampleItemView> list(Integer limit) {
        return repository.list(ListLimit.of(limit)).stream().map(SampleItemView::of).toList();
    }

    public void delete(UUID id) {
        if (!repository.delete(id)) {
            throw notFound();
        }
    }

    private SampleItem load(UUID id) {
        return repository.findById(id).orElseThrow(SampleItemService::notFound);
    }

    private static NotFoundException notFound() {
        return new NotFoundException("SampleItem não encontrado.");
    }
}
