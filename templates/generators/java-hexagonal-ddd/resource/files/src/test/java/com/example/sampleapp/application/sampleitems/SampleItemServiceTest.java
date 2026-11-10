package com.example.sampleapp.application.sampleitems;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.example.sampleapp.application.shared.NotFoundException;
import com.example.sampleapp.domain.sampleitem.SampleItem;
import java.time.Clock;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import org.junit.jupiter.api.Test;

class SampleItemServiceTest {

    private static final class FakeSampleItemRepository implements SampleItemRepository {

        private final Map<UUID, SampleItem> saved = new HashMap<>();

        @Override
        public void save(SampleItem sampleItem) {
            saved.put(sampleItem.id(), sampleItem);
        }

        @Override
        public Optional<SampleItem> findById(UUID id) {
            return Optional.ofNullable(saved.get(id));
        }

        @Override
        public List<SampleItem> list(int limit) {
            return saved.values().stream().limit(limit).toList();
        }

        @Override
        public boolean delete(UUID id) {
            return saved.remove(id) != null;
        }
    }

    private final SampleItemService service = new SampleItemService(new FakeSampleItemRepository(), Clock.systemUTC());

    @Test
    void createRenameAndDelete() {
        UUID id = service.create("Primeiro");
        service.rename(id, "Segundo");

        assertThat(service.get(id).name()).isEqualTo("Segundo");
        assertThat(service.list(null)).hasSize(1);

        service.delete(id);
        assertThatThrownBy(() -> service.get(id)).isInstanceOf(NotFoundException.class);
        assertThatThrownBy(() -> service.delete(id)).isInstanceOf(NotFoundException.class);
    }
}
