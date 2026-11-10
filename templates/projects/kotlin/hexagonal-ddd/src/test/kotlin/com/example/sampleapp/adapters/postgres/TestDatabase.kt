package com.example.sampleapp.adapters.postgres

import org.junit.jupiter.api.Assumptions.assumeTrue

object TestDatabase {
    private val connected by lazy { Database.connect(DatabaseSettings.fromUrl(System.getenv("DATABASE_URL"))) }

    fun connectOrSkip(): Database {
        val url = System.getenv("DATABASE_URL")
        assumeTrue(!url.isNullOrBlank(), "DATABASE_URL not set: romanctl up sample-app --no-run creates the database and the .env")
        return connected
    }
}
