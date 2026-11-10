package com.example.sampleapp.adapters.http;

import java.util.UUID;
import org.springframework.http.HttpMethod;
import org.springframework.http.HttpStatusCode;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.client.RestClient;

final class ApiClient {

    private final RestClient rest;

    ApiClient(int port) {
        this.rest = RestClient.builder()
                .baseUrl("http://localhost:" + port)
                .defaultStatusHandler(HttpStatusCode::isError, (request, response) -> {
                })
                .build();
    }

    ResponseEntity<String> send(HttpMethod method, String path, String json) {
        RestClient.RequestBodySpec request = rest.method(method).uri(path).contentType(MediaType.APPLICATION_JSON);
        if (json != null) {
            request.body(json);
        }
        return request.retrieve().toEntity(String.class);
    }

    <T> T get(String path, Class<T> type) {
        return rest.get().uri(path).retrieve().body(type);
    }

    static UUID createdId(ResponseEntity<?> response) {
        String location = response.getHeaders().getLocation().getPath();
        return UUID.fromString(location.substring(location.lastIndexOf('/') + 1));
    }
}
