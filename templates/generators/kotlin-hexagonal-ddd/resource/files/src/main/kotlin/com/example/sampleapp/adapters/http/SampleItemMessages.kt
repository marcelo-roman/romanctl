package com.example.sampleapp.adapters.http

import com.example.sampleapp.application.sampleitems.SampleItemView
import kotlinx.serialization.Serializable

@Serializable
data class SampleItemRequest(val name: String = "")

@Serializable
data class SampleItemResponse(val id: String, val name: String, val createdAt: String) {
    companion object {
        fun of(view: SampleItemView): SampleItemResponse =
            SampleItemResponse(view.id.toString(), view.name, view.createdAt.toString())
    }
}
