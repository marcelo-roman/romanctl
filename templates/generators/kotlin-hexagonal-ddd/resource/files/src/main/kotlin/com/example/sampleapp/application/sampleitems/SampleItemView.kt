package com.example.sampleapp.application.sampleitems

import com.example.sampleapp.domain.sampleitem.SampleItem
import java.time.Instant
import java.util.UUID

data class SampleItemView(val id: UUID, val name: String, val createdAt: Instant) {
    companion object {
        internal fun of(sampleItem: SampleItem): SampleItemView =
            SampleItemView(sampleItem.id, sampleItem.name, sampleItem.createdAt)
    }
}
