package com.example.sampleapp.adapters.postgres;

import com.example.sampleapp.application.sampleitems.SampleItemRepository;
import com.example.sampleapp.domain.sampleitem.SampleItem;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;

@Repository
@ConditionalOnProperty(name = "app.persistence", havingValue = "postgres")
public class PostgresSampleItemRepository implements SampleItemRepository {

    private static final String SELECT = "SELECT id, name, created_at FROM sample_items";

    private final JdbcClient jdbc;

    public PostgresSampleItemRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    @Override
    public void save(SampleItem sampleItem) {
        jdbc.sql("""
                INSERT INTO sample_items (id, name, created_at) VALUES (:id, :name, :createdAt)
                ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
                """)
                .param("id", sampleItem.id())
                .param("name", sampleItem.name())
                .param("createdAt", OffsetDateTime.ofInstant(sampleItem.createdAt(), ZoneOffset.UTC))
                .update();
    }

    @Override
    public Optional<SampleItem> findById(UUID id) {
        return jdbc.sql(SELECT + " WHERE id = :id").param("id", id).query(PostgresSampleItemRepository::row).optional();
    }

    @Override
    public List<SampleItem> list(int limit) {
        return jdbc.sql(SELECT + " ORDER BY created_at DESC LIMIT :limit")
                .param("limit", limit)
                .query(PostgresSampleItemRepository::row)
                .list();
    }

    @Override
    public boolean delete(UUID id) {
        return jdbc.sql("DELETE FROM sample_items WHERE id = :id").param("id", id).update() > 0;
    }

    private static SampleItem row(ResultSet rs, int rowNum) throws SQLException {
        return SampleItem.restore(rs.getObject("id", UUID.class), rs.getString("name"),
                rs.getObject("created_at", OffsetDateTime.class).toInstant());
    }
}
