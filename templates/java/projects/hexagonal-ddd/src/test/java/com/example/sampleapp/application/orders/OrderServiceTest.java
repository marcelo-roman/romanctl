package com.example.sampleapp.application.orders;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.example.sampleapp.application.shared.NotFoundException;
import com.example.sampleapp.domain.order.Order;
import com.example.sampleapp.domain.order.OrderPlaced;
import com.example.sampleapp.domain.shared.DomainEvent;
import java.math.BigDecimal;
import java.time.Clock;
import java.time.Instant;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import org.junit.jupiter.api.Test;

class OrderServiceTest {

    private static final class FakeOrderRepository implements OrderRepository {

        private final Map<UUID, Order> saved = new HashMap<>();

        @Override
        public void save(Order order) {
            saved.put(order.id(), order);
        }

        @Override
        public Optional<Order> findById(UUID id) {
            return Optional.ofNullable(saved.get(id));
        }

        @Override
        public List<Order> list(int limit) {
            return saved.values().stream().limit(limit).toList();
        }
    }

    private final List<DomainEvent> published = new ArrayList<>();
    private final Clock clock = Clock.fixed(Instant.parse("2026-01-01T00:00:00Z"), ZoneOffset.UTC);
    private final OrderService service = new OrderService(new FakeOrderRepository(), published::addAll, clock);

    @Test
    void placingPublishesEvent() {
        UUID id = service.create("Ana");
        service.addItem(id, new AddItem("Caneta", 3, new BigDecimal("2.50"), null));
        service.place(id);

        OrderView view = service.get(id);
        assertThat(view.status()).isEqualTo("Placed");
        assertThat(view.total()).isEqualByComparingTo("7.50");
        assertThat(view.currency()).isEqualTo("BRL");
        assertThat(view.createdAt()).isEqualTo(clock.instant());
        assertThat(published).singleElement().isInstanceOf(OrderPlaced.class);
    }

    @Test
    void unknownOrderIsNotFound() {
        assertThatThrownBy(() -> service.place(UUID.randomUUID()))
                .isInstanceOf(NotFoundException.class)
                .hasMessage("Pedido não encontrado.");
    }

    @Test
    void listReturnsViews() {
        service.create("Ana");
        service.create("Bia");

        assertThat(service.list(null)).hasSize(2);
        assertThat(service.list(1)).hasSize(1);
    }
}
