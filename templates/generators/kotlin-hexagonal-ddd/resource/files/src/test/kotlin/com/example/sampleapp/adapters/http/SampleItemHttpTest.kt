package com.example.sampleapp.adapters.http

import io.ktor.http.HttpMethod
import io.ktor.http.HttpStatusCode
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class SampleItemHttpTest {
    @Test
    fun crud() = apiTest {
        val created = client.send(HttpMethod.Post, "/sample-items", """{"name":"Primeiro"}""")
        assertEquals(HttpStatusCode.Created, created.status)
        val id = created.createdId()

        assertEquals(HttpStatusCode.NoContent, client.send(HttpMethod.Put, "/sample-items/$id", """{"name":"Segundo"}""").status)
        val loaded = client.send(HttpMethod.Get, "/sample-items/$id").json().jsonObject
        assertEquals("Segundo", loaded.getValue("name").jsonPrimitive.content)
        val list = client.send(HttpMethod.Get, "/sample-items?limit=10").json().jsonArray
        assertTrue(list.any { it.jsonObject.getValue("id").jsonPrimitive.content == id })

        assertEquals(HttpStatusCode.NoContent, client.send(HttpMethod.Delete, "/sample-items/$id").status)
        assertEquals(HttpStatusCode.NotFound, client.send(HttpMethod.Get, "/sample-items/$id").status)
    }

    @Test
    fun `blank name is rejected`() = apiTest {
        assertEquals(HttpStatusCode.UnprocessableEntity, client.send(HttpMethod.Post, "/sample-items", """{"name":" "}""").status)
    }
}
