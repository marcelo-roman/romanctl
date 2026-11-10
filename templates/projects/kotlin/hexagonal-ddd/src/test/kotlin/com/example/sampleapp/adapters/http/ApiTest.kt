package com.example.sampleapp.adapters.http

import io.ktor.client.HttpClient
import io.ktor.client.request.request
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpMethod
import io.ktor.http.contentType
import io.ktor.server.testing.ApplicationTestBuilder
import io.ktor.server.testing.testApplication
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive

fun apiTest(block: suspend ApplicationTestBuilder.() -> Unit) = testApplication {
    application { api(inMemoryServices()) }
    block()
}

suspend fun HttpClient.send(method: HttpMethod, path: String, body: String? = null): HttpResponse = request(path) {
    this.method = method
    if (body != null) {
        contentType(ContentType.Application.Json)
        setBody(body)
    }
}

suspend fun HttpResponse.json(): JsonElement = Json.parseToJsonElement(bodyAsText())

suspend fun HttpResponse.createdId(): String = json().jsonObject.getValue("id").jsonPrimitive.content
