package com.example.sampleapp.domain.shared

import java.util.UUID

object Identifier {
    private val format = Regex("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

    fun parse(value: String): UUID {
        if (!format.matches(value)) {
            throw BusinessRuleException("Identificador inválido.")
        }
        return UUID.fromString(value)
    }
}
