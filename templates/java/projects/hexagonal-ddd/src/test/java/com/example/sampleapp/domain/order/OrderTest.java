package com.example.sampleapp.domain.order;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.example.sampleapp.domain.shared.BusinessRuleException;
import com.example.sampleapp.domain.shared.DomainEvent;
import java.math.BigDecimal;
import java.time.Instant;
import java.util.List;
import org.junit.jupiter.api.Test;

class OrderTest {

    private static final Instant NOW = Instant.parse("2026-01-01T12:00:00Z");

    private static Money brl(String amount) {
        return Money.of(new BigDecimal(amount), "brl");
    }

    @Test
    void createRequiresCustomer() {
        assertThatThrownBy(() -> Order.create("  ", NOW)).isInstanceOf(BusinessRuleException.class);
    }

    @Test
    void createTrimsCustomerAndStartsAsDraft() {
        Order order = Order.create("  Ana ", NOW);

        assertThat(order.customerName()).isEqualTo("Ana");
        assertThat(order.status()).isEqualTo(OrderStatus.DRAFT);
        assertThat(order.total()).isEqualTo(Money.zero("BRL"));
    }

    @Test
    void totalSumsItems() {
        Order order = Order.create("Ana", NOW);
        order.addItem("Caneta", 2, brl("3.50"));
        order.addItem("Caderno", 1, brl("12"));

        assertThat(order.total()).isEqualTo(brl("19.00"));
    }

    @Test
    void itemsMustShareCurrency() {
        Order order = Order.create("Ana", NOW);
        order.addItem("Caneta", 1, brl("1"));

        assertThatThrownBy(() -> order.addItem("Pen", 1, Money.of(BigDecimal.ONE, "USD")))
                .isInstanceOf(BusinessRuleException.class)
                .hasMessage("Todos os itens do pedido devem usar a mesma moeda.");
    }

    @Test
    void itemRequiresProductAndPositiveQuantity() {
        Order order = Order.create("Ana", NOW);

        assertThatThrownBy(() -> order.addItem(" ", 1, brl("1"))).isInstanceOf(BusinessRuleException.class);
        assertThatThrownBy(() -> order.addItem("Caneta", 0, brl("1"))).isInstanceOf(BusinessRuleException.class);
    }

    @Test
    void placeRaisesEventOnce() {
        Order order = Order.create("Ana", NOW);
        assertThatThrownBy(() -> order.place(NOW)).isInstanceOf(BusinessRuleException.class);

        order.addItem("Caneta", 1, brl("1"));
        order.place(NOW);

        List<DomainEvent> events = order.pullEvents();
        assertThat(events).singleElement().isInstanceOfSatisfying(OrderPlaced.class,
                placed -> assertThat(placed.total()).isEqualByComparingTo("1"));
        assertThat(order.pullEvents()).isEmpty();
        assertThat(order.status()).isEqualTo(OrderStatus.PLACED);
        assertThatThrownBy(() -> order.place(NOW)).isInstanceOf(BusinessRuleException.class);
        assertThatThrownBy(() -> order.addItem("Lápis", 1, brl("1"))).isInstanceOf(BusinessRuleException.class);
    }

    @Test
    void moneyRejectsNegativeAmountAndInvalidCurrency() {
        assertThatThrownBy(() -> Money.of(new BigDecimal("-1"), "BRL")).isInstanceOf(BusinessRuleException.class);
        assertThatThrownBy(() -> Money.of(BigDecimal.ONE, "REAL")).isInstanceOf(BusinessRuleException.class);
        assertThatThrownBy(() -> Money.of(null, "BRL")).isInstanceOf(BusinessRuleException.class);
    }

    @Test
    void moneyDefaultsToBrl() {
        assertThat(Money.of(BigDecimal.TEN, null).currency()).isEqualTo("BRL");
    }
}
