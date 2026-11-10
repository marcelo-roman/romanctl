package com.example.sampleapp.adapters.http;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.sampleapp.application.orders.OrderView;
import java.util.UUID;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.HttpMethod;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;

@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT,
        properties = {"app.persistence=memory", "spring.sql.init.mode=never"})
class OrderHttpTest {

    @Value("${local.server.port}")
    private int port;

    private ApiClient api;

    @BeforeEach
    void setUp() {
        api = new ApiClient(port);
    }

    @Test
    void orderLifecycle() {
        ResponseEntity<String> created = api.send(HttpMethod.POST, "/orders", "{\"customerName\":\"Ana\"}");
        assertThat(created.getStatusCode()).isEqualTo(HttpStatus.CREATED);
        UUID id = ApiClient.createdId(created);

        ResponseEntity<String> added = api.send(HttpMethod.POST, "/orders/" + id + "/items",
                "{\"productName\":\"Caneta\",\"quantity\":2,\"unitPrice\":3.5}");
        assertThat(added.getStatusCode()).isEqualTo(HttpStatus.NO_CONTENT);
        assertThat(api.send(HttpMethod.POST, "/orders/" + id + "/place", null).getStatusCode()).isEqualTo(HttpStatus.NO_CONTENT);

        OrderView order = api.get("/orders/" + id, OrderView.class);
        assertThat(order.status()).isEqualTo("Placed");
        assertThat(order.total()).isEqualByComparingTo("7");
        assertThat(order.items()).hasSize(1);
        assertThat(api.get("/orders?limit=10", OrderView[].class)).extracting(OrderView::id).contains(id);
    }

    @Test
    void errorsBecomeProblems() {
        ResponseEntity<String> rule = api.send(HttpMethod.POST, "/orders", "{\"customerName\":\"\"}");
        assertThat(rule.getStatusCode().value()).isEqualTo(422);
        assertThat(rule.getBody()).contains("O nome do cliente é obrigatório.");

        ResponseEntity<String> missing = api.send(HttpMethod.GET, "/orders/" + UUID.randomUUID(), null);
        assertThat(missing.getStatusCode()).isEqualTo(HttpStatus.NOT_FOUND);
        assertThat(missing.getHeaders().getContentType()).hasToString("application/problem+json");

        assertThat(api.send(HttpMethod.POST, "/orders", "{").getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
    }

    @Test
    void healthIsOk() {
        assertThat(api.send(HttpMethod.GET, "/health", null).getBody()).contains("ok");
    }
}
