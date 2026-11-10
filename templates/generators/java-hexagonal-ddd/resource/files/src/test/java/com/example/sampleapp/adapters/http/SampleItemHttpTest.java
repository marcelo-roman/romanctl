package com.example.sampleapp.adapters.http;

import static org.assertj.core.api.Assertions.assertThat;

import com.example.sampleapp.application.sampleitems.SampleItemView;
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
class SampleItemHttpTest {

    @Value("${local.server.port}")
    private int port;

    private ApiClient api;

    @BeforeEach
    void setUp() {
        api = new ApiClient(port);
    }

    @Test
    void crud() {
        ResponseEntity<String> created = api.send(HttpMethod.POST, "/sample-items", "{\"name\":\"Primeiro\"}");
        assertThat(created.getStatusCode()).isEqualTo(HttpStatus.CREATED);
        UUID id = ApiClient.createdId(created);

        assertThat(api.send(HttpMethod.PUT, "/sample-items/" + id, "{\"name\":\"Segundo\"}").getStatusCode())
                .isEqualTo(HttpStatus.NO_CONTENT);
        assertThat(api.get("/sample-items/" + id, SampleItemView.class).name()).isEqualTo("Segundo");
        assertThat(api.get("/sample-items?limit=10", SampleItemView[].class)).extracting(SampleItemView::id).contains(id);

        assertThat(api.send(HttpMethod.DELETE, "/sample-items/" + id, null).getStatusCode()).isEqualTo(HttpStatus.NO_CONTENT);
        assertThat(api.send(HttpMethod.GET, "/sample-items/" + id, null).getStatusCode()).isEqualTo(HttpStatus.NOT_FOUND);
    }

    @Test
    void blankNameIsRejected() {
        assertThat(api.send(HttpMethod.POST, "/sample-items", "{\"name\":\" \"}").getStatusCode().value()).isEqualTo(422);
    }
}
