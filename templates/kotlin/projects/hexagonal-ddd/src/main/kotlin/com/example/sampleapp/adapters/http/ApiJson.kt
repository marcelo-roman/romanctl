package com.example.sampleapp.adapters.http

import kotlinx.serialization.json.Json

object ApiJson {
    val format = Json { ignoreUnknownKeys = true }
}
