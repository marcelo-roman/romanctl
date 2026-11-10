package com.example.sampleapp.application.sampleitems

import com.example.sampleapp.application.shared.NotFoundException
import com.example.sampleapp.domain.sampleitem.SampleItem
import java.time.Clock
import java.time.Instant
import java.time.ZoneOffset
import java.util.UUID
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class SampleItemServiceTest {
    private class FakeSampleItemRepository : SampleItemRepository {
        val saved = mutableMapOf<UUID, SampleItem>()

        override fun save(sampleItem: SampleItem) {
            saved[sampleItem.id] = sampleItem
        }

        override fun findById(id: UUID): SampleItem? = saved[id]

        override fun list(limit: Int): List<SampleItem> = saved.values.take(limit)

        override fun delete(id: UUID): Boolean = saved.remove(id) != null
    }

    private val clock = Clock.fixed(Instant.parse("2026-01-01T10:00:00Z"), ZoneOffset.UTC)
    private val service = SampleItemService(FakeSampleItemRepository(), clock)

    @Test
    fun `create, rename, list and delete`() {
        val id = service.create("Primeiro").toString()

        service.rename(id, "Segundo")

        assertEquals("Segundo", service.get(id).name)
        assertEquals(clock.instant(), service.get(id).createdAt)
        assertEquals(1, service.list(null).size)
        service.delete(id)
        assertFailsWith<NotFoundException> { service.get(id) }
    }

    @Test
    fun `deleting an unknown item is not found`() {
        assertFailsWith<NotFoundException> { service.delete(UUID.randomUUID().toString()) }
    }
}
