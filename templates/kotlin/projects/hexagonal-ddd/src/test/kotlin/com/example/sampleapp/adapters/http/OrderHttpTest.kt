package com.example.sampleapp.adapters.http

import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpMethod
import io.ktor.http.HttpStatusCode
import io.ktor.http.contentType
import kotlinx.serialization.json.int
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class OrderHttpTest {
    @Test
    fun `health answers ok`() = apiTest {
        val response = client.send(HttpMethod.Get, "/health")

        assertEquals(HttpStatusCode.OK, response.status)
        assertEquals("ok", response.json().jsonObject.getValue("status").jsonPrimitive.content)
    }

    @Test
    fun `order flow`() = apiTest {
        val created = client.send(HttpMethod.Post, "/orders", """{"customerName":"Ana"}""")
        assertEquals(HttpStatusCode.Created, created.status)
        val id = created.createdId()
        assertEquals("/orders/$id", created.headers[HttpHeaders.Location])

        val added = client.send(HttpMethod.Post, "/orders/$id/items", """{"productName":"Caneta","quantity":2,"unitPrice":3.5}""")
        assertEquals(HttpStatusCode.NoContent, added.status)
        assertEquals(HttpStatusCode.NoContent, client.send(HttpMethod.Post, "/orders/$id/place").status)

        val order = client.send(HttpMethod.Get, "/orders/$id").json().jsonObject
        assertEquals("Placed", order.getValue("status").jsonPrimitive.content)
        assertEquals("Ana", order.getValue("customerName").jsonPrimitive.content)
        assertEquals("7.00", order.getValue("total").jsonPrimitive.content)
        assertEquals("BRL", order.getValue("currency").jsonPrimitive.content)
        assertTrue(order.containsKey("createdAt"))
        val item = order.getValue("items").jsonArray.single().jsonObject
        assertEquals("3.50", item.getValue("unitPrice").jsonPrimitive.content)
        assertEquals("7.00", item.getValue("subtotal").jsonPrimitive.content)

        val list = client.send(HttpMethod.Get, "/orders?limit=10").json().jsonArray
        assertTrue(list.any { it.jsonObject.getValue("id").jsonPrimitive.content == id })
    }

    @Test
    fun `rule violation is a 422 problem`() = apiTest {
        val id = client.send(HttpMethod.Post, "/orders", """{"customerName":"Ana"}""").createdId()

        val response = client.send(HttpMethod.Post, "/orders/$id/place")

        assertEquals(HttpStatusCode.UnprocessableEntity, response.status)
        assertEquals(ContentType("application", "problem+json"), response.contentType()?.withoutParameters())
        val problem = response.json().jsonObject
        assertEquals(422, problem.getValue("status").jsonPrimitive.int)
        assertEquals("Regra de negócio violada", problem.getValue("title").jsonPrimitive.content)
        assertEquals("Um pedido sem itens não pode ser realizado.", problem.getValue("detail").jsonPrimitive.content)
    }

    @Test
    fun `blank customer is a 422`() = apiTest {
        assertEquals(HttpStatusCode.UnprocessableEntity, client.send(HttpMethod.Post, "/orders", """{"customerName":" "}""").status)
    }

    @Test
    fun `unknown order is a 404`() = apiTest {
        val response = client.send(HttpMethod.Get, "/orders/00000000-0000-4000-8000-000000000000")

        assertEquals(HttpStatusCode.NotFound, response.status)
        assertEquals("Pedido não encontrado.", response.json().jsonObject.getValue("detail").jsonPrimitive.content)
    }

    @Test
    fun `malformed id is a 422`() = apiTest {
        val response = client.send(HttpMethod.Get, "/orders/abc")

        assertEquals(HttpStatusCode.UnprocessableEntity, response.status)
        assertEquals("Identificador inválido.", response.json().jsonObject.getValue("detail").jsonPrimitive.content)
    }

    @Test
    fun `invalid body is a 400`() = apiTest {
        val response = client.send(HttpMethod.Post, "/orders", "{")

        assertEquals(HttpStatusCode.BadRequest, response.status, response.bodyAsText())
        assertEquals(400, response.json().jsonObject.getValue("status").jsonPrimitive.int)
    }
}
