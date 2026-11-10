package com.example.sampleapp.adapters.postgres;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.sampleapp.domain.order.Money;
import com.example.sampleapp.domain.order.Order;
import com.example.sampleapp.domain.order.OrderStatus;
import java.math.BigDecimal;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.UUID;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.condition.EnabledIfEnvironmentVariable;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

@EnabledIfEnvironmentVariable(named = "DATABASE_URL", matches = ".+",
        disabledReason = "DATABASE_URL not set: romanctl up sample-app --no-run creates the database and the .env")
@SpringBootTest(properties = "app.persistence=postgres")
class PostgresOrderRepositoryTest {

    @Autowired
    private PostgresOrderRepository repository;

    @Test
    void roundTrip() {
        Order order = Order.create("Ana", Instant.now().truncatedTo(ChronoUnit.MICROS));
        order.addItem("Caneta", 2, Money.of(new BigDecimal("3.50"), "BRL"));
        repository.save(order);
        order.place(Instant.now());
        repository.save(order);

        Order loaded = repository.findById(order.id()).orElseThrow();

        assertThat(loaded.customerName()).isEqualTo("Ana");
        assertThat(loaded.status()).isEqualTo(OrderStatus.PLACED);
        assertThat(loaded.createdAt()).isEqualTo(order.createdAt());
        assertThat(loaded.items()).hasSize(1);
        assertThat(loaded.total()).isEqualTo(Money.of(new BigDecimal("7"), "BRL"));
        assertThat(repository.list(200)).extracting(Order::id).contains(order.id());
        assertThat(repository.findById(UUID.randomUUID())).isEmpty();
    }
}
