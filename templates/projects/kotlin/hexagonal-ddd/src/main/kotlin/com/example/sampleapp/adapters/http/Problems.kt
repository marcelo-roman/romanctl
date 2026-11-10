package com.example.sampleapp.adapters.http

import com.example.sampleapp.application.shared.NotFoundException
import com.example.sampleapp.domain.shared.BusinessRuleException
import io.ktor.http.ContentType
import io.ktor.http.HttpStatusCode
import io.ktor.server.application.ApplicationCall
import io.ktor.server.application.log
import io.ktor.server.plugins.BadRequestException
import io.ktor.server.plugins.ContentTransformationException
import io.ktor.server.plugins.statuspages.StatusPagesConfig
import io.ktor.server.response.respondText
import kotlinx.serialization.Serializable

@Serializable
data class Problem(val status: Int, val title: String, val detail: String)

private val problemJson = ContentType("application", "problem+json")

fun StatusPagesConfig.problems() {
    exception<BusinessRuleException> { call, cause ->
        call.problem(HttpStatusCode.UnprocessableEntity, "Regra de negócio violada", cause.message.orEmpty())
    }
    exception<NotFoundException> { call, cause ->
        call.problem(HttpStatusCode.NotFound, "Não encontrado", cause.message.orEmpty())
    }
    exception<BadRequestException> { call, cause ->
        call.problem(HttpStatusCode.BadRequest, "Requisição inválida", cause.message.orEmpty())
    }
    exception<ContentTransformationException> { call, cause ->
        call.problem(HttpStatusCode.BadRequest, "Requisição inválida", cause.message.orEmpty())
    }
    exception<Throwable> { call, cause ->
        call.application.log.error("request failed", cause)
        call.problem(HttpStatusCode.InternalServerError, "Erro interno", "Erro inesperado.")
    }
}

private suspend fun ApplicationCall.problem(status: HttpStatusCode, title: String, detail: String) {
    val body = ApiJson.format.encodeToString(Problem.serializer(), Problem(status.value, title, detail))
    respondText(body, problemJson, status)
}
