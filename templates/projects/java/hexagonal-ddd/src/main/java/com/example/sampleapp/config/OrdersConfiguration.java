package com.example.sampleapp.config;

import com.example.sampleapp.application.orders.OrderRepository;
import com.example.sampleapp.application.orders.OrderService;
import com.example.sampleapp.application.shared.EventPublisher;
import java.time.Clock;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class OrdersConfiguration {

    @Bean
    public OrderService orderService(OrderRepository repository, EventPublisher events, Clock clock) {
        return new OrderService(repository, events, clock);
    }
}
