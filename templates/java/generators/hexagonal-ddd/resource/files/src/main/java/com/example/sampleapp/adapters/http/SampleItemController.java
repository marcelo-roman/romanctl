package com.example.sampleapp.adapters.http;

import com.example.sampleapp.application.sampleitems.SampleItemService;
import com.example.sampleapp.application.sampleitems.SampleItemView;
import java.net.URI;
import java.util.List;
import java.util.UUID;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/sample-items")
public class SampleItemController {

    public record SampleItemRequest(String name) {
    }

    private final SampleItemService sampleItems;

    public SampleItemController(SampleItemService sampleItems) {
        this.sampleItems = sampleItems;
    }

    @PostMapping
    public ResponseEntity<CreatedResponse> create(@RequestBody SampleItemRequest request) {
        UUID id = sampleItems.create(request.name());
        return ResponseEntity.created(URI.create("/sample-items/" + id)).body(new CreatedResponse(id));
    }

    @GetMapping
    public List<SampleItemView> list(@RequestParam(required = false) Integer limit) {
        return sampleItems.list(limit);
    }

    @GetMapping("/{id}")
    public SampleItemView get(@PathVariable UUID id) {
        return sampleItems.get(id);
    }

    @PutMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void rename(@PathVariable UUID id, @RequestBody SampleItemRequest request) {
        sampleItems.rename(id, request.name());
    }

    @DeleteMapping("/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void delete(@PathVariable UUID id) {
        sampleItems.delete(id);
    }
}
