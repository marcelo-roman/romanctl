package com.example.sampleapp.config;

import com.example.sampleapp.application.sampleitems.SampleItemRepository;
import com.example.sampleapp.application.sampleitems.SampleItemService;
import java.time.Clock;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class SampleItemsConfiguration {

    @Bean
    public SampleItemService sampleItemService(SampleItemRepository repository, Clock clock) {
        return new SampleItemService(repository, clock);
    }
}
