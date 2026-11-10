package com.example.sampleapp.adapters.postgres

import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariDataSource
import java.sql.Connection

class Database private constructor(private val dataSource: HikariDataSource) : AutoCloseable {
    fun <T> read(block: (Connection) -> T): T = dataSource.connection.use(block)

    fun <T> transaction(block: (Connection) -> T): T = dataSource.connection.use { connection ->
        connection.autoCommit = false
        try {
            val result = block(connection)
            connection.commit()
            result
        } catch (failure: Throwable) {
            connection.rollback()
            throw failure
        }
    }

    override fun close() = dataSource.close()

    companion object {
        fun connect(settings: DatabaseSettings): Database {
            val config = HikariConfig().apply {
                jdbcUrl = settings.jdbcUrl
                username = settings.user
                password = settings.password
                maximumPoolSize = 10
            }
            val database = Database(HikariDataSource(config))
            Schema.create(database)
            return database
        }
    }
}
