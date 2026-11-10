package com.example.sampleapp.adapters.http;

import com.example.sampleapp.application.orders.AddItem;
import com.example.sampleapp.application.orders.OrderService;
import com.example.sampleapp.application.orders.OrderView;
import java.math.BigDecimal;
import java.net.URI;
import java.util.List;
import java.util.UUID;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/orders")
public class OrderController {

    public record CreateOrderRequest(String customerName) {
    }

    public record AddItemRequest(String productName, int quantity, BigDecimal unitPrice, String currency) {
    }

    private final OrderService orders;

    public OrderController(OrderService orders) {
        this.orders = orders;
    }

    @PostMapping
    public ResponseEntity<CreatedResponse> create(@RequestBody CreateOrderRequest request) {
        UUID id = orders.create(request.customerName());
        return ResponseEntity.created(URI.create("/orders/" + id)).body(new CreatedResponse(id));
    }

    @PostMapping("/{id}/items")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void addItem(@PathVariable UUID id, @RequestBody AddItemRequest request) {
        orders.addItem(id, new AddItem(request.productName(), request.quantity(), request.unitPrice(), request.currency()));
    }

    @PostMapping("/{id}/place")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void place(@PathVariable UUID id) {
        orders.place(id);
    }

    @GetMapping("/{id}")
    public OrderView get(@PathVariable UUID id) {
        return orders.get(id);
    }

    @GetMapping
    public List<OrderView> list(@RequestParam(required = false) Integer limit) {
        return orders.list(limit);
    }
}
