package com.example.sampleapp.adapters.postgres

import kotlin.test.Test
import kotlin.test.assertEquals

class DatabaseSettingsTest {
    @Test
    fun `converts DATABASE_URL to JDBC`() {
        val settings = DatabaseSettings.fromUrl("postgres://app:p%40ss@db.local:6543/sample_app?sslmode=disable")

        assertEquals("jdbc:postgresql://db.local:6543/sample_app?sslmode=disable", settings.jdbcUrl)
        assertEquals("app", settings.user)
        assertEquals("p@ss", settings.password)
    }

    @Test
    fun `falls back to the local database`() {
        val settings = DatabaseSettings.from(emptyMap())

        assertEquals("jdbc:postgresql://127.0.0.1:5432/sample_app", settings.jdbcUrl)
        assertEquals("postgres", settings.user)
    }
}
