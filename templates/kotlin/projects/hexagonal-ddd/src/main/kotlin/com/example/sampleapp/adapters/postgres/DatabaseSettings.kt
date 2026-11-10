package com.example.sampleapp.adapters.postgres

import java.net.URI
import java.net.URLDecoder
import java.nio.charset.StandardCharsets

data class DatabaseSettings(val jdbcUrl: String, val user: String, val password: String) {
    companion object {
        const val DEFAULT_URL = "postgres://postgres:postgres@127.0.0.1:5432/sample_app"
        private const val DEFAULT_PORT = 5432
        private val schemes = setOf("postgres", "postgresql")

        fun from(environment: Map<String, String>): DatabaseSettings {
            val url = environment["DATABASE_URL"]
            if (url.isNullOrBlank()) {
                return fromUrl(DEFAULT_URL)
            }
            return fromUrl(url)
        }

        fun fromUrl(url: String): DatabaseSettings {
            val uri = URI(url)
            require(uri.scheme in schemes) { "DATABASE_URL must start with postgres://" }
            val port = if (uri.port == -1) DEFAULT_PORT else uri.port
            val query = uri.rawQuery?.let { "?$it" }.orEmpty()
            val credentials = uri.rawUserInfo.orEmpty().split(":", limit = 2)
            return DatabaseSettings(
                jdbcUrl = "jdbc:postgresql://${uri.host}:$port${uri.rawPath}$query",
                user = decode(credentials[0]),
                password = decode(credentials.getOrElse(1) { "" }),
            )
        }

        private fun decode(value: String): String = URLDecoder.decode(value, StandardCharsets.UTF_8)
    }
}
