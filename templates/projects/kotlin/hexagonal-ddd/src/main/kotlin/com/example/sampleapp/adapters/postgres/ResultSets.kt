package com.example.sampleapp.adapters.postgres

import java.sql.ResultSet

internal fun <T> ResultSet.collect(map: (ResultSet) -> T): List<T> {
    val out = mutableListOf<T>()
    while (next()) {
        out.add(map(this))
    }
    return out
}
