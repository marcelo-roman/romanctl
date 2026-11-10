package com.example.sampleapp.adapters.postgres

import com.example.sampleapp.application.sampleitems.SampleItemRepository
import com.example.sampleapp.domain.sampleitem.SampleItem
import java.sql.ResultSet
import java.time.OffsetDateTime
import java.time.ZoneOffset
import java.util.UUID

class PostgresSampleItemRepository(private val database: Database) : SampleItemRepository {
    override fun save(sampleItem: SampleItem) {
        database.transaction { connection ->
            connection.prepareStatement(UPSERT).use {
                it.setObject(1, sampleItem.id)
                it.setString(2, sampleItem.name)
                it.setObject(3, OffsetDateTime.ofInstant(sampleItem.createdAt, ZoneOffset.UTC))
                it.executeUpdate()
            }
        }
    }

    override fun findById(id: UUID): SampleItem? = database.read { connection ->
        connection.prepareStatement("$SELECT WHERE id = ?").use { statement ->
            statement.setObject(1, id)
            statement.executeQuery().use { it.collect(::row) }
        }
    }.firstOrNull()

    override fun list(limit: Int): List<SampleItem> = database.read { connection ->
        connection.prepareStatement("$SELECT ORDER BY created_at DESC LIMIT ?").use { statement ->
            statement.setInt(1, limit)
            statement.executeQuery().use { it.collect(::row) }
        }
    }

    override fun delete(id: UUID): Boolean = database.transaction { connection ->
        connection.prepareStatement("DELETE FROM sample_items WHERE id = ?").use {
            it.setObject(1, id)
            it.executeUpdate() > 0
        }
    }

    private fun row(rs: ResultSet): SampleItem = SampleItem.restore(
        rs.getObject("id", UUID::class.java),
        rs.getString("name"),
        rs.getObject("created_at", OffsetDateTime::class.java).toInstant(),
    )

    private companion object {
        const val SELECT = "SELECT id, name, created_at FROM sample_items"
        const val UPSERT = """
            INSERT INTO sample_items (id, name, created_at) VALUES (?, ?, ?)
            ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
        """
    }
}
