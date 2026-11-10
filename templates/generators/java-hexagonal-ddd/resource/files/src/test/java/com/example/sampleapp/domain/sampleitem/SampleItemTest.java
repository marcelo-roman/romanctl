package com.example.sampleapp.domain.sampleitem;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.example.sampleapp.domain.shared.BusinessRuleException;
import java.time.Instant;
import org.junit.jupiter.api.Test;

class SampleItemTest {

    @Test
    void createRequiresName() {
        assertThatThrownBy(() -> SampleItem.create(" ", Instant.now())).isInstanceOf(BusinessRuleException.class);
        assertThat(SampleItem.create("  Primeiro ", Instant.now()).name()).isEqualTo("Primeiro");
    }

    @Test
    void renameRequiresName() {
        SampleItem sampleItem = SampleItem.create("Primeiro", Instant.now());

        assertThatThrownBy(() -> sampleItem.rename("")).isInstanceOf(BusinessRuleException.class);
        sampleItem.rename("Segundo");
        assertThat(sampleItem.name()).isEqualTo("Segundo");
    }
}
