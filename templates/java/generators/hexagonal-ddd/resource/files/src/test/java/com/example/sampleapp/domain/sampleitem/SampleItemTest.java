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
        assertThat(SampleItem.create("  First ", Instant.now()).name()).isEqualTo("First");
    }

    @Test
    void renameRequiresName() {
        SampleItem sampleItem = SampleItem.create("First", Instant.now());

        assertThatThrownBy(() -> sampleItem.rename("")).isInstanceOf(BusinessRuleException.class);
        sampleItem.rename("Second");
        assertThat(sampleItem.name()).isEqualTo("Second");
    }
}
