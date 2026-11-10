package com.example.sampleapp.adapters.postgres;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.sampleapp.domain.sampleitem.SampleItem;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.condition.EnabledIfEnvironmentVariable;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

@EnabledIfEnvironmentVariable(named = "DATABASE_URL", matches = ".+",
        disabledReason = "DATABASE_URL not set: romanctl up --no-run creates the database and the .env")
@SpringBootTest(properties = "app.persistence=postgres")
class PostgresSampleItemRepositoryTest {

    @Autowired
    private PostgresSampleItemRepository repository;

    @Test
    void roundTrip() {
        SampleItem sampleItem = SampleItem.create("Primeiro", Instant.now().truncatedTo(ChronoUnit.MICROS));
        repository.save(sampleItem);
        sampleItem.rename("Segundo");
        repository.save(sampleItem);

        SampleItem loaded = repository.findById(sampleItem.id()).orElseThrow();
        assertThat(loaded.name()).isEqualTo("Segundo");
        assertThat(loaded.createdAt()).isEqualTo(sampleItem.createdAt());
        assertThat(repository.list(200)).extracting(SampleItem::id).contains(sampleItem.id());

        assertThat(repository.delete(sampleItem.id())).isTrue();
        assertThat(repository.findById(sampleItem.id())).isEmpty();
    }
}
