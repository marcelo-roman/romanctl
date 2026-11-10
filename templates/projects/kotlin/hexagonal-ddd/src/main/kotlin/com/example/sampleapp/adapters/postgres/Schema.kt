package com.example.sampleapp.adapters.postgres

object Schema {
    private const val RESOURCE = "/schema.sql"

    fun create(database: Database) {
        val ddl = requireNotNull(Schema::class.java.getResource(RESOURCE)) { "$RESOURCE not found" }.readText()
        database.transaction { connection ->
            connection.createStatement().use { statement ->
                statement.execute("SELECT pg_advisory_xact_lock(hashtext('schema'))")
                statement.execute(ddl)
            }
        }
    }
}
