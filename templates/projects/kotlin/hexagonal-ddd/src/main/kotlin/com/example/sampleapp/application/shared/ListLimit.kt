package com.example.sampleapp.application.shared

object ListLimit {
    const val DEFAULT = 50
    const val MAX = 200

    fun of(requested: Int?): Int {
        if (requested == null || requested <= 0) {
            return DEFAULT
        }
        return minOf(requested, MAX)
    }
}
