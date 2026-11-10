package com.example.sampleapp.adapters.postgres

import com.example.sampleapp.domain.sampleitem.SampleItem
import java.time.Instant
import java.time.temporal.ChronoUnit
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class PostgresSampleItemRepositoryTest {
    @Test
    fun `round trip`() {
        val repository = PostgresSampleItemRepository(TestDatabase.connectOrSkip())
        val now = Instant.now().truncatedTo(ChronoUnit.MICROS)
        val sampleItem = SampleItem.create("Primeiro", now)
        repository.save(sampleItem)
        sampleItem.rename("Segundo")
        repository.save(sampleItem)

        val loaded = assertNotNull(repository.findById(sampleItem.id))

        assertEquals("Segundo", loaded.name)
        assertEquals(now, loaded.createdAt)
        assertTrue(repository.list(200).any { it.id == sampleItem.id })
        assertTrue(repository.delete(sampleItem.id))
        assertFalse(repository.delete(sampleItem.id))
        assertNull(repository.findById(sampleItem.id))
    }
}
