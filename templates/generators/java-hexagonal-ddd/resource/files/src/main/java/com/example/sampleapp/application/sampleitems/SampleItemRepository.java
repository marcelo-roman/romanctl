package com.example.sampleapp.application.sampleitems;

import com.example.sampleapp.domain.sampleitem.SampleItem;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

public interface SampleItemRepository {

    void save(SampleItem sampleItem);

    Optional<SampleItem> findById(UUID id);

    List<SampleItem> list(int limit);

    boolean delete(UUID id);
}
