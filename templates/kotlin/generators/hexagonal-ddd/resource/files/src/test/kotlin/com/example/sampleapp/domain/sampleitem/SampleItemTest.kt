package com.example.sampleapp.domain.sampleitem

import com.example.sampleapp.domain.shared.BusinessRuleException
import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class SampleItemTest {
    private val now = Instant.parse("2026-01-01T10:00:00Z")

    @Test
    fun `create trims the name`() {
        val sampleItem = SampleItem.create("  Primeiro  ", now)

        assertEquals("Primeiro", sampleItem.name)
        assertEquals(now, sampleItem.createdAt)
    }

    @Test
    fun `name is required`() {
        val error = assertFailsWith<BusinessRuleException> { SampleItem.create(" ", now) }

        assertEquals("O nome é obrigatório.", error.message)
    }

    @Test
    fun `rename validates the new name`() {
        val sampleItem = SampleItem.create("Primeiro", now)

        sampleItem.rename("Segundo")

        assertEquals("Segundo", sampleItem.name)
        assertFailsWith<BusinessRuleException> { sampleItem.rename("") }
    }
}
