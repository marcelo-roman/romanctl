package com.example.sampleapp.domain.sampleitem

import com.example.sampleapp.domain.shared.BusinessRuleException
import java.time.Instant
import java.util.UUID

class SampleItem private constructor(val id: UUID, name: String, val createdAt: Instant) {
    var name: String = name
        private set

    fun rename(name: String) {
        this.name = validName(name)
    }

    companion object {
        fun create(name: String, now: Instant): SampleItem = SampleItem(UUID.randomUUID(), validName(name), now)

        fun restore(id: UUID, name: String, createdAt: Instant): SampleItem = SampleItem(id, name, createdAt)

        private fun validName(name: String): String {
            if (name.isBlank()) {
                throw BusinessRuleException("O nome é obrigatório.")
            }
            return name.trim()
        }
    }
}
